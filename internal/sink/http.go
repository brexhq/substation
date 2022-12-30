package sink

import (
	"context"
	"fmt"
	"os"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/http"
	"github.com/brexhq/substation/internal/json"
)

var httpClient http.HTTP

// http sinks data to an HTTP(S) URL.
type _http struct {
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

// Send sinks a channel of encapsulated data with the sink.
func (sink *_http) Send(ctx context.Context, ch *config.Channel) error {
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

			if len(sink.Headers) > 0 {
				for _, header := range sink.Headers {
					headers = append(headers, http.Header{
						Key:   header.Key,
						Value: header.Value,
					})
				}
			}

			if sink.HeadersKey != "" {
				h := capsule.Get(sink.HeadersKey).Array()
				for _, header := range h {
					for k, v := range header.Map() {
						headers = append(headers, http.Header{
							Key:   k,
							Value: v.String(),
						})
					}
				}
			}

			_, err := httpClient.Post(ctx, sink.URL, string(capsule.Data()), headers...)
			if err != nil {
				// Post err returns metadata
				return fmt.Errorf("sink: http: %v", err)
			}
		}
	}

	return nil
}
