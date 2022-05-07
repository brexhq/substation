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
		maps values from the JSON object (Key) to the HTTP header (Header)
		defaults to no headers

The sink uses this Jsonnet configuration:
	{
		type: 'http',
		settings: {
			url: 'foo.com/bar',
			headers: [
				key: 'foo',
				header: 'X-FOO',
			],
		},
	}
*/
type HTTP struct {
	URL     string `mapstructure:"url"`
	Headers []struct {
		Key    string `mapstructure:"key"`
		Header string `mapstructure:"header"`
	} `mapstructure:"headers"`
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

				for _, h := range sink.Headers {
					v := json.Get(data, h.Header).String()
					headers = append(headers, http.Header{
						Key:   h.Key,
						Value: v,
					})
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
