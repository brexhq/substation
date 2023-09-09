package transform

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/brexhq/substation/config"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/http"
	"github.com/brexhq/substation/internal/secrets"
	"github.com/brexhq/substation/message"
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

func (c *sendHTTPConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *sendHTTPConfig) Validate() error {
	if c.URL == "" {
		return fmt.Errorf("url: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

type sendHTTP struct {
	conf sendHTTPConfig

	// client is safe for concurrent use.
	client http.HTTP
}

func newSendHTTP(_ context.Context, cfg config.Config) (*sendHTTP, error) {
	conf := sendHTTPConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: new_send_http: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: new_send_http: %v", err)
	}

	tf := sendHTTP{
		conf: conf,
	}

	tf.client.Setup()
	if _, ok := os.LookupEnv("AWS_XRAY_DAEMON_ADDRESS"); ok {
		tf.client.EnableXRay()
	}

	return &tf, nil
}

func (tf *sendHTTP) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	var headers []http.Header

	if json.Valid(msg.Data()) {
		headers = append(headers, http.Header{
			Key:   "Content-Type",
			Value: "application/json",
		})
	}

	for _, hdr := range tf.conf.Headers {
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

	if tf.conf.HeadersKey != "" {
		h := msg.GetValue(tf.conf.HeadersKey).Array()
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
	url, err := secrets.Interpolate(ctx, tf.conf.URL)
	if err != nil {
		return nil, fmt.Errorf("transform: send_http: %v", err)
	}

	resp, err := tf.client.Post(ctx, url, string(msg.Data()), headers...)
	if err != nil {
		// Post errors return metadata.
		return nil, fmt.Errorf("transform: send_http: %v", err)
	}

	//nolint:errcheck // Response body is discarded to avoid resource leaks.
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()

	return []*message.Message{msg}, nil
}

func (tf *sendHTTP) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}

func (*sendHTTP) Close(context.Context) error {
	return nil
}
