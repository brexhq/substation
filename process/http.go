//go:build !wasm

package process

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	gohttp "net/http"
	"strings"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/http"
	"github.com/brexhq/substation/internal/secrets"
)

// httpInterp is used for interpolating data into URLs.
const httpInterp = `${data}`

var httpClient http.HTTP

// http processes data by retrieving a payload from an HTTP(S) URL. The HTTP client
// used by the processor uses an exponential retry strategy that makes up to four requests
// and does not wait more than 30 seconds for each retry, which may have significant impact
// on end-to-end data processing latency. If Substation is running in AWS Lambda with
// Kinesis, then this latency can be mitigated by increasing the parallelization factor
// of the Lambda (https://docs.aws.amazon.com/lambda/latest/dg/with-kinesis.html).
//
// This processor supports the data and object handling patterns.
type procHTTP struct {
	process
	Options procHTTPOptions `json:"options"`

	headers []http.Header
}

type procHTTPOptions struct {
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

// Closes resources opened by the processor.
func (p procHTTP) Close(context.Context) error {
	return nil
}

// Create a new HTTP processor.
func newProcHTTP(ctx context.Context, cfg config.Config) (p procHTTP, err error) {
	if err = config.Decode(cfg.Settings, &p); err != nil {
		return procHTTP{}, err
	}

	p.operator, err = condition.NewOperator(ctx, p.Condition)
	if err != nil {
		return procHTTP{}, err
	}

	// error early if required options are missing
	if p.Options.URL == "" {
		return procHTTP{}, fmt.Errorf("process: http: option url: %v", errors.ErrMissingRequiredOption)
	}

	if p.Options.Method == "POST" && p.Options.BodyKey == "" {
		return procHTTP{}, fmt.Errorf("process: http: options body_key: %v", errors.ErrMissingRequiredOption)
	}

	if !httpClient.IsEnabled() {
		httpClient.Setup()
	}

	for _, hdr := range p.Options.Headers {
		// retrieve secret and interpolate with header value
		v, err := secrets.Interpolate(ctx, hdr.Value)
		if err != nil {
			return procHTTP{}, fmt.Errorf("process: http: %v", err)
		}

		p.headers = append(p.headers, http.Header{
			Key:   hdr.Key,
			Value: v,
		})
	}

	return p, nil
}

// Stream processes a pipeline of capsules with the processor.
func (p procHTTP) Stream(ctx context.Context, in, out *config.Channel) error {
	return streamApply(ctx, in, out, p)
}

// Batch processes one or more capsules with the processor.
func (p procHTTP) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	return batchApply(ctx, capsules, p)
}

// Apply processes a capsule with the processor.
func (p procHTTP) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	if ok, err := p.operator.Operate(ctx, capsule); err != nil {
		return capsule, fmt.Errorf("process: http: %v", err)
	} else if !ok {
		return capsule, nil
	}

	// the URL can exist in three states:
	//
	// - no interpolation, the URL is unchanged
	//
	// - object-based interpolation, the URL is interpolated
	// using the object handling pattern.
	//
	// - data-based interpolation, the URL is interpolated
	// using the data handling pattern.
	//
	// the URL is always interpolated with the substring ${data}
	url := p.Options.URL
	if strings.Contains(url, httpInterp) {
		if p.Key != "" {
			url = strings.Replace(url, httpInterp, capsule.Get(p.Key).String(), 1)
		} else {
			url = strings.Replace(url, httpInterp, string(capsule.Data()), 1)
		}
	}

	// retrieve secret and interpolate with URL
	url, err := secrets.Interpolate(ctx, url)
	if err != nil {
		return capsule, fmt.Errorf("process: http: %v", err)
	}

	switch p.Options.Method {
	// POST only supports the object handling pattern
	case gohttp.MethodPost:
		// BodyKey is a requirement, otherwise there is no payload to send
		if p.Options.BodyKey == "" {
			return capsule, fmt.Errorf("process: http: options %+v: %v", p.Options, errors.ErrMissingRequiredOption)
		}

		body := capsule.Get(p.Options.BodyKey).String()
		resp, err := httpClient.Post(ctx, url, body, p.headers...)
		if err != nil && p.IgnoreErrors {
			return capsule, nil
		} else if err != nil {
			return capsule, fmt.Errorf("process: http: %v", err)
		}
		defer resp.Body.Close()

		res, err := parseResponse(resp)
		if err != nil {
			return capsule, fmt.Errorf("process: http: %v", err)
		}

		// if SetKey exists, then the response body is written into the capsule,
		// but otherwise the response is not stored and the capsule is returned
		// as-is
		if p.SetKey != "" {
			if err := capsule.Set(p.SetKey, res); err != nil {
				return capsule, fmt.Errorf("process: http: %v", err)
			}
			return capsule, nil
		}

		return capsule, nil
	// GET must be the last condition and fallthrough since it is the default Method
	case gohttp.MethodGet:
		fallthrough
	default:
		resp, err := httpClient.Get(ctx, url, p.headers...)
		if err != nil && p.IgnoreErrors {
			return capsule, nil
		} else if err != nil {
			return capsule, fmt.Errorf("process: http: %v", err)
		}
		defer resp.Body.Close()

		res, err := parseResponse(resp)
		if err != nil {
			return capsule, fmt.Errorf("process: http: %v", err)
		}

		// object processing
		if p.SetKey != "" {
			if err := capsule.Set(p.SetKey, res); err != nil {
				return capsule, fmt.Errorf("process: http: %v", err)
			}
			return capsule, nil
		}

		// data processing
		capsule.SetData(res)
		return capsule, nil
	}
}

func parseResponse(resp *gohttp.Response) ([]byte, error) {
	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("process: http: %v", err)
	}

	dst := &bytes.Buffer{}
	if json.Valid(buf) {
		// compact converts a multi-line object into a single-line object.
		if err := json.Compact(dst, buf); err != nil {
			return nil, fmt.Errorf("process: http: %v", err)
		}
	} else {
		dst = bytes.NewBuffer(buf)
	}

	return dst.Bytes(), nil
}
