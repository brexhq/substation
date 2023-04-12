package sink

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/http"
	"github.com/brexhq/substation/internal/json"
	"github.com/brexhq/substation/internal/secrets"
)

var httpClient http.HTTP

// http sinks data to an HTTP(S) URL.
type sinkHTTP struct {
	// URL is the HTTP(S) endpoint that data is sent to.
	URL string `json:"url"`
	// Headers are an array of objects that contain HTTP headers sent in the request.
	//
	// This is optional and has no default.
	Headers []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	} `json:"headers"`
	// HeadersKey retrieves a value from an object that contains one or
	// more objects containing HTTP headers sent in the request. If Headers
	// is used, then both are merged together.
	//
	// This is optional and has no default.
	HeadersKey string `json:"headers_key"`
}

// Create a new HTTP sink.
func newSinkHTTP(cfg config.Config) (s sinkHTTP, err error) {
	if err = config.Decode(cfg.Settings, &s); err != nil {
		return sinkHTTP{}, err
	}

	if s.URL == "" {
		return sinkHTTP{}, fmt.Errorf("sink: http: URL: %v", errors.ErrMissingRequiredOption)
	}

	return s, nil
}

// Send sinks a channel of encapsulated data with the sink.
func (s sinkHTTP) Send(ctx context.Context, ch *config.Channel) error {
	if !httpClient.IsEnabled() {
		httpClient.Setup()
		if _, ok := os.LookupEnv("AWS_XRAY_DAEMON_ADDRESS"); ok {
			httpClient.EnableXRay()
		}
	}

	for capsule := range ch.C {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			var headers []http.Header

			if json.Valid(capsule.Data()) {
				headers = append(headers, http.Header{
					Key:   "Content-Type",
					Value: "application/json",
				})
			}

			for _, hdr := range s.Headers {
				// retrieve secret and interpolate with header value
				v, err := secrets.Interpolate(ctx, hdr.Value)
				if err != nil {
					return fmt.Errorf("sink: http: %v", err)
				}

				headers = append(headers, http.Header{
					Key:   hdr.Key,
					Value: v,
				})
			}

			if s.HeadersKey != "" {
				h := capsule.Get(s.HeadersKey).Array()
				for _, header := range h {
					for k, v := range header.Map() {
						headers = append(headers, http.Header{
							Key:   k,
							Value: v.String(),
						})
					}
				}
			}

			// retrieve secret and interpolate with URL
			url, err := secrets.Interpolate(ctx, s.URL)
			if err != nil {
				return fmt.Errorf("sink: http: %v", err)
			}

			resp, err := httpClient.Post(ctx, url, string(capsule.Data()), headers...)
			if err != nil {
				// Post err returns metadata
				return fmt.Errorf("sink: http: %v", err)
			}

			//nolint:errcheck // response body is discarded to avoid resource leaks
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
	}

	return nil
}
