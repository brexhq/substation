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
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/http"
	"github.com/brexhq/substation/internal/secrets"
	"github.com/brexhq/substation/message"
)

// modHTTPInterp is used for interpolating data into URLs.
const modHTTPInterp = `${data}`

type modHTTPConfig struct {
	Object configObject `json:"object"`

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

type modHTTP struct {
	conf modHTTPConfig

	// client is safe for concurrent use.
	client  http.HTTP
	headers []http.Header
}

func newModHTTP(ctx context.Context, cfg config.Config) (*modHTTP, error) {
	conf := modHTTPConfig{}
	if err := iconfig.Decode(cfg.Settings, &conf); err != nil {
		return nil, fmt.Errorf("transform: new_mod_http: %v", err)
	}

	// Validate required options.
	if conf.Object.Key == "" && conf.Object.SetKey != "" {
		return nil, fmt.Errorf("transform: new_mod_http: object_key: %v", errors.ErrMissingRequiredOption)
	}

	if conf.Object.Key != "" && conf.Object.SetKey == "" {
		return nil, fmt.Errorf("transform: new_mod_http: object_set_key: %v", errors.ErrMissingRequiredOption)
	}

	if conf.URL == "" {
		return nil, fmt.Errorf("transform: new_mod_http: url: %v", errors.ErrMissingRequiredOption)
	}

	if conf.Method == "POST" && conf.BodyKey == "" {
		return nil, fmt.Errorf("transform: new_mod_http: body_key: %v", errors.ErrMissingRequiredOption)
	}

	tf := modHTTP{
		conf: conf,
	}

	tf.client.Setup()
	for _, hdr := range conf.Headers {
		// Retrieve secret and interpolate with header value.
		v, err := secrets.Interpolate(ctx, hdr.Value)
		if err != nil {
			return nil, fmt.Errorf("transform: new_mod_http: %v", err)
		}

		tf.headers = append(tf.headers, http.Header{
			Key:   hdr.Key,
			Value: v,
		})
	}

	return &tf, nil
}

func (*modHTTP) Close(context.Context) error {
	return nil
}

func (tf *modHTTP) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	// Skip interrupt messages.
	if msg.IsControl() {
		return []*message.Message{msg}, nil
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
	url := tf.conf.URL
	if strings.Contains(url, modHTTPInterp) {
		if tf.conf.Object.Key != "" {
			url = strings.ReplaceAll(url, modHTTPInterp, msg.GetObject(tf.conf.Object.Key).String())
		} else {
			url = strings.ReplaceAll(url, modHTTPInterp, string(msg.Data()))
		}
	}

	// Retrieve secret and interpolate with URL
	url, err := secrets.Interpolate(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("transform: mod_http: %v", err)
	}

	switch tf.conf.Method {
	// POST only supports the object handling pattern.
	case gohttp.MethodPost:
		body := msg.GetObject(tf.conf.BodyKey).String()

		// resp.Body is closed by parseResponse.
		resp, err := tf.client.Post(ctx, url, body, tf.headers...)

		// If ErrorOnFailure is configured, then errors are returned,
		// but otherwise the message is returned as-is.
		if err != nil && tf.conf.ErrorOnFailure {
			return nil, fmt.Errorf("transform: mod_http: %v", err)
		} else if err != nil {
			//nolint: nilerr // err is configurable.
			return []*message.Message{msg}, nil
		}

		res, err := tf.parseResponse(resp)
		if err != nil {
			return nil, fmt.Errorf("transform: mod_http: %v", err)
		}

		// If SetKey exists, then the response body is written into the message,
		// but otherwise the response is not stored and the message is returned
		// as-is.
		if tf.conf.Object.SetKey != "" {
			if err := msg.SetObject(tf.conf.Object.SetKey, res); err != nil {
				return nil, fmt.Errorf("transform: mod_http: %v", err)
			}
		}

		return []*message.Message{msg}, nil

	// GET must be the last condition and fallthrough since it is the default Method
	case gohttp.MethodGet:
		fallthrough
	default:
		// resp.Body is closed by parseResponse.
		resp, err := tf.client.Get(ctx, url, tf.headers...)

		// If ErrorOnFailure is configured, then errors are returned,
		// but otherwise the message is returned as-is.
		if err != nil && tf.conf.ErrorOnFailure {
			return nil, fmt.Errorf("transform: mod_http: %v", err)
		} else if err != nil {
			//nolint: nilerr // err is configurable.
			return []*message.Message{msg}, nil
		}

		res, err := tf.parseResponse(resp)
		if err != nil {
			return nil, fmt.Errorf("transform: mod_http: %v", err)
		}

		if tf.conf.Object.SetKey != "" {
			if err := msg.SetObject(tf.conf.Object.SetKey, res); err != nil {
				return nil, fmt.Errorf("transform: mod_http: %v", err)
			}

			return []*message.Message{msg}, nil
		}

		finMsg := message.New().SetData(res).SetMetadata(msg.Metadata())
		return []*message.Message{finMsg}, nil
	}
}

func (tf *modHTTP) parseResponse(resp *gohttp.Response) ([]byte, error) {
	defer resp.Body.Close()

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	dst := &bytes.Buffer{}
	if json.Valid(buf) {
		// Compact converts a multi-line object into a single-line object.
		if err := json.Compact(dst, buf); err != nil {
			return nil, err
		}
	} else {
		dst = bytes.NewBuffer(buf)
	}

	return dst.Bytes(), nil
}
