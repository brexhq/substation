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

/*
HTTP sinks JSON data to an HTTP(S) endpoint.

The sink has these settings:
	URL:
		HTTP(S) endpoint that data is sent to
	Headers (optional):
		contains configured maps that represent HTTP headers to be sent in the HTTP request
		defaults to no headers
	HeadersKey (optional):
		JSON key-value that contains maps that represent HTTP headers to be sent in the HTTP request
		This key can be a single map or an array of maps:
			[
				{
					"FOO": "bar",
				},
				{
					"BAZ": "qux",
				}
			]

When loaded with a factory, the sink uses this JSON configuration:
	{
		"type": "http",
		"settings": {
			"url": "foo.com/bar"
		}
	}
*/
type HTTP struct {
	URL     string `json:"url"`
	Headers []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	} `json:"headers"`
	HeadersKey string `json:"headers_key"`
}

// Send sinks a channel of encapsulated data with the HTTP sink.
func (sink *HTTP) Send(ctx context.Context, ch *config.Channel) error {
	if !httpClient.IsEnabled() {
		httpClient.Setup()
		if _, ok := os.LookupEnv("AWS_XRAY_DAEMON_ADDRESS"); ok {
			httpClient.EnableXRay()
		}
	}

	for cap := range ch.C {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			var headers []http.Header

			if json.Valid(cap.GetData()) {
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
				h := cap.Get(sink.HeadersKey).Array()
				for _, header := range h {
					for k, v := range header.Map() {
						headers = append(headers, http.Header{
							Key:   k,
							Value: v.String(),
						})
					}
				}
			}

			_, err := httpClient.Post(ctx, sink.URL, string(cap.GetData()), headers...)
			if err != nil {
				// Post err returns metadata
				return fmt.Errorf("sink http: %v", err)
			}
		}
	}

	return nil
}
