package sink

import (
	"context"
	"fmt"
	"os"

	"github.com/brexhq/substation/internal/http"
	"github.com/brexhq/substation/internal/json"
)

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

The sink uses this Jsonnet configuration:
	{
		type: 'http',
		settings: {
			url: 'foo.com/bar',
			headers_key: 'foo',
		},
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

var httpClient http.HTTP

// Send sinks a channel of bytes with the HTTP sink.
func (sink *HTTP) Send(ctx context.Context, ch chan []byte, kill chan struct{}) error {
	if !httpClient.IsEnabled() {
		httpClient.Setup()
		if _, ok := os.LookupEnv("AWS_XRAY_DAEMON_ADDRESS"); ok {
			httpClient.EnableXRay()
		}
	}

	for data := range ch {
		select {
		case <-kill:
			return nil
		default:
			var headers []http.Header

			if json.Valid(data) {
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
				h := json.Get(data, sink.HeadersKey).Array()
				for _, header := range h {
					for k, v := range header.Map() {
						headers = append(headers, http.Header{
							Key:   k,
							Value: v.String(),
						})
					}
				}
			}

			_, err := httpClient.Post(ctx, sink.URL, string(data), headers...)
			if err != nil {
				return fmt.Errorf("err failed to POST to URL %s: %v", sink.URL, err)
			}
		}
	}

	return nil
}
