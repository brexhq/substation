package transform

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/brexhq/substation/config"
	_config "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/http"
	"github.com/brexhq/substation/internal/json"
	"github.com/brexhq/substation/internal/secrets"
	mess "github.com/brexhq/substation/message"
)

type sendHTTPConfig struct {
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

type sendHTTP struct {
	conf sendHTTPConfig

	// client is safe for concurrent use.
	client http.HTTP
}

func newSendHTTP(_ context.Context, cfg config.Config) (*sendHTTP, error) {
	conf := sendHTTPConfig{}
	if err := _config.Decode(cfg.Settings, &conf); err != nil {
		return nil, err
	}

	if conf.URL == "" {
		return nil, fmt.Errorf("send: http: URL: %v", errors.ErrMissingRequiredOption)
	}

	send := sendHTTP{
		conf: conf,
	}

	send.client.Setup()
	if _, ok := os.LookupEnv("AWS_XRAY_DAEMON_ADDRESS"); ok {
		send.client.EnableXRay()
	}

	return &send, nil
}

func (*sendHTTP) Close(context.Context) error {
	return nil
}

func (t *sendHTTP) Transform(ctx context.Context, messages ...*mess.Message) ([]*mess.Message, error) {
	for _, message := range messages {
		if message.IsControl() {
			continue
		}

		var headers []http.Header

		if json.Valid(message.Data()) {
			headers = append(headers, http.Header{
				Key:   "Content-Type",
				Value: "application/json",
			})
		}

		for _, hdr := range t.conf.Headers {
			// Retrieve secret and interpolate with header value.
			v, err := secrets.Interpolate(ctx, hdr.Value)
			if err != nil {
				return nil, fmt.Errorf("transform: send_http: %v", err)
			}

			headers = append(headers, http.Header{
				Key:   hdr.Key,
				Value: v,
			})
		}

		if t.conf.HeadersKey != "" {
			h := message.Get(t.conf.HeadersKey).Array()
			for _, header := range h {
				for k, v := range header.Map() {
					headers = append(headers, http.Header{
						Key:   k,
						Value: v.String(),
					})
				}
			}
		}

		// Retrieve secret and interpolate with URL.
		url, err := secrets.Interpolate(ctx, t.conf.URL)
		if err != nil {
			return nil, fmt.Errorf("transform: send_http: %v", err)
		}

		resp, err := t.client.Post(ctx, url, string(message.Data()), headers...)
		if err != nil {
			// Post errors return metadata.
			return nil, fmt.Errorf("transform: send_http: %v", err)
		}

		//nolint:errcheck // Response body is discarded to avoid resource leaks.
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}

	return messages, nil
}
