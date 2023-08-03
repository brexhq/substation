//go:build !wasm

package transform

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	gohttp "net/http"
	"strings"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/http"
	"github.com/brexhq/substation/internal/secrets"
	mess "github.com/brexhq/substation/message"
)

// procHTTPInterp is used for interpolating data into URLs.
const procHTTPInterp = `${data}`

type procHTTPConfig struct {
	// Key retrieves a value from an object for processing.
	//
	// This is optional for transforms that support processing non-object data.
	Key string `json:"key"`
	// SetKey inserts a processed value into an object.
	//
	// This is optional for transforms that support processing non-object data.
	SetKey string `json:"set_key"`
	// ErrorOnFailure determines whether an error is returned during processing.
	//
	// This is optional and defaults to false.
	ErrorOnFailure bool `json:"error_on_failure"`
	// Method is the HTTP method used in the call.
	//
	// Must be one of:
	//
	// - GET
	//
	// - POST
	//
	// Defaults to GET.
	Method string `json:"method"`
	// URL is the HTTP(S) endpoint that data is retrieved from.
	//
	// If the substring ${data} is in the URL, then the URL is interpolated with
	// data (either the value from Key or the raw data). URLs may be optionally
	// interpolated with secrets (e.g., ${SECRETS_ENV:FOO}).
	URL string `json:"url"`
	// Headers are an array of objects that contain HTTP headers sent in the request.
	// Values may be optionally interpolated with secrets (e.g., ${SECRETS_ENV:FOO}).
	//
	// This is optional and has no default.
	Headers []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	} `json:"headers"`
	// BodyKey retrieves a value from an object that is used as the message body.
	// This is only used in HTTP requests that send payloads to the server.
	//
	// This is optional and has no default.
	BodyKey string `json:"body_key"`
}

type procHTTP struct {
	conf     procHTTPConfig
	isObject bool

	// client is safe for concurrent use.
	client  http.HTTP
	headers []http.Header
}

func (*procHTTP) Close(context.Context) error {
	return nil
}

func newProcHTTP(ctx context.Context, cfg config.Config) (*procHTTP, error) {
	conf := procHTTPConfig{}
	if err := config.Decode(cfg.Settings, &conf); err != nil {
		return nil, err
	}

	// Validate required options.
	if (conf.Key != "" && conf.SetKey == "") ||
		(conf.Key == "" && conf.SetKey != "") {
		return nil, fmt.Errorf("transform: proc_http: key %s set_key %s: %v", conf.Key, conf.SetKey, errInvalidDataPattern)
	}

	if conf.URL == "" {
		return nil, fmt.Errorf("transform: proc_http: url: %v", errors.ErrMissingRequiredOption)
	}

	if conf.Method == "POST" && conf.BodyKey == "" {
		return nil, fmt.Errorf("transform: proc_http: body_key: %v", errors.ErrMissingRequiredOption)
	}

	proc := procHTTP{
		conf:     conf,
		isObject: conf.Key != "" && conf.SetKey != "",
	}

	proc.client.Setup()
	for _, hdr := range conf.Headers {
		// Retrieve secret and interpolate with header value.
		v, err := secrets.Interpolate(ctx, hdr.Value)
		if err != nil {
			return nil, fmt.Errorf("transform: proc_http: %v", err)
		}

		proc.headers = append(proc.headers, http.Header{
			Key:   hdr.Key,
			Value: v,
		})
	}

	return &proc, nil
}

//nolint: gocognit // Ignore cognitive complexity.
func (t *procHTTP) Transform(ctx context.Context, messages ...*mess.Message) ([]*mess.Message, error) {
	var output []*mess.Message

	for _, message := range messages {
		// Skip control messages.
		if message.IsControl() {
			output = append(output, message)
			continue
		}

		// The URL can exist in three states:
		//
		// - No interpolation, the URL is unchanged.
		//
		// - Object-based interpolation, the URL is interpolated
		// using the object handling pattern.
		//
		// - Data-based interpolation, the URL is interpolated
		// using the data handling pattern.
		//
		// The URL is always interpolated with the substring ${data}.
		url := t.conf.URL
		if strings.Contains(url, procHTTPInterp) {
			if t.conf.Key != "" {
				url = strings.ReplaceAll(url, procHTTPInterp, message.Get(t.conf.Key).String())
			} else {
				url = strings.ReplaceAll(url, procHTTPInterp, string(message.Data()))
			}
		}

		// Retrieve secret and interpolate with URL
		url, err := secrets.Interpolate(ctx, url)
		if err != nil {
			return nil, fmt.Errorf("transform: proc_http: %v", err)
		}

		switch t.conf.Method {
		// POST only supports the object handling pattern.
		case gohttp.MethodPost:
			body := message.Get(t.conf.BodyKey).String()
			resp, err := t.client.Post(ctx, url, body, t.headers...)
			if err != nil && t.conf.ErrorOnFailure {
				output = append(output, message)
				continue
			} else if err != nil {
				return nil, fmt.Errorf("transform: proc_http: %v", err)
			}
			defer resp.Body.Close()

			res, err := parseResponse(resp)
			if err != nil {
				return nil, fmt.Errorf("transform: proc_http: %v", err)
			}

			// If SetKey exists, then the response body is written into the message,
			// but otherwise the response is not stored and the message is returned
			// as-is.
			if t.conf.SetKey != "" {
				if err := message.Set(t.conf.SetKey, res); err != nil {
					return nil, fmt.Errorf("transform: proc_http: %v", err)
				}

				output = append(output, message)
				continue
			}

			output = append(output, message)
			continue
		// GET must be the last condition and fallthrough since it is the default Method
		case gohttp.MethodGet:
			fallthrough
		default:
			resp, err := t.client.Get(ctx, url, t.headers...)
			if err != nil && t.conf.ErrorOnFailure {
				output = append(output, message)
				continue
			} else if err != nil {
				return nil, fmt.Errorf("transform: proc_http: %v", err)
			}
			defer resp.Body.Close()

			res, err := parseResponse(resp)
			if err != nil {
				return nil, fmt.Errorf("transform: proc_http: %v", err)
			}

			if t.conf.SetKey != "" {
				if err := message.Set(t.conf.SetKey, res); err != nil {
					return nil, fmt.Errorf("transform: proc_http: %v", err)
				}

				output = append(output, message)
				continue
			}

			msg, err := mess.New(
				mess.SetData(res),
				mess.SetMetadata(message.Metadata()),
			)
			if err != nil {
				return nil, fmt.Errorf("process: dns: %v", err)
			}

			output = append(output, msg)
		}
	}

	return output, nil
}

func parseResponse(resp *gohttp.Response) ([]byte, error) {
	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("transform: proc_http: %v", err)
	}

	dst := &bytes.Buffer{}
	if json.Valid(buf) {
		// Compact converts a multi-line object into a single-line object.
		if err := json.Compact(dst, buf); err != nil {
			return nil, fmt.Errorf("transform: proc_http: %v", err)
		}
	} else {
		dst = bytes.NewBuffer(buf)
	}

	return dst.Bytes(), nil
}
