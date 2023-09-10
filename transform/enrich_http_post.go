//go:build !wasm

package transform

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/brexhq/substation/config"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/http"
	"github.com/brexhq/substation/internal/secrets"
	"github.com/brexhq/substation/message"
)

type enrichHTTPPostConfig struct {
	Object iconfig.Object `json:"object"`

	// URL is the HTTP(S) endpoint that data is retrieved from.
	//
	// If the substring ${data} is in the URL, then the URL is interpolated with
	// data (either the value from Object.Key or the raw data). URLs may be optionally
	// interpolated with secrets (e.g., ${SECRETS_ENV:FOO}).
	URL string `json:"url"`
	// BodyKey retrieves a value from an object that is used as the message body.
	BodyKey string `json:"body_key"`
	// Headers are an array of objects that contain HTTP headers sent in the request.
	// Values may be optionally interpolated with secrets (e.g., ${SECRETS_ENV:FOO}).
	//
	// This is optional and has no default.
	Headers []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	} `json:"headers"`
}

func (c *enrichHTTPPostConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *enrichHTTPPostConfig) Validate() error {
	if c.Object.Key == "" && c.Object.SetKey != "" {
		return fmt.Errorf("object_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.Object.Key != "" && c.Object.SetKey == "" {
		return fmt.Errorf("object_set_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.URL == "" {
		return fmt.Errorf("url: %v", errors.ErrMissingRequiredOption)
	}

	if c.BodyKey == "" {
		return fmt.Errorf("body_key: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

func newEnrichHTTPPost(ctx context.Context, cfg config.Config) (*enrichHTTPPost, error) {
	conf := enrichHTTPPostConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: new_enrich_http_post: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: new_enrich_http_post: %v", err)
	}

	tf := enrichHTTPPost{
		conf: conf,
	}

	tf.client.Setup()
	for _, hdr := range conf.Headers {
		// Retrieve secret and interpolate with header value.
		v, err := secrets.Interpolate(ctx, hdr.Value)
		if err != nil {
			return nil, fmt.Errorf("transform: new_enrich_http_post: %v", err)
		}

		tf.headers = append(tf.headers, http.Header{
			Key:   hdr.Key,
			Value: v,
		})
	}

	return &tf, nil
}

type enrichHTTPPost struct {
	conf enrichHTTPPostConfig

	// client is safe for concurrent use.
	client  http.HTTP
	headers []http.Header
}

func (tf *enrichHTTPPost) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
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
	if strings.Contains(url, enrichHTTPInterp) {
		if tf.conf.Object.Key != "" {
			value := msg.GetValue(tf.conf.Object.Key)
			url = strings.ReplaceAll(url, enrichHTTPInterp, value.String())
		} else {
			url = strings.ReplaceAll(url, enrichHTTPInterp, string(msg.Data()))
		}
	}

	// Retrieve secret and interpolate with URL
	url, err := secrets.Interpolate(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("transform: enrich_http_post: %v", err)
	}

	bodyValue := msg.GetValue(tf.conf.BodyKey)

	// resp.Body is closed by enrichHTTPParseResponse.
	resp, err := tf.client.Post(ctx, url, bodyValue.String(), tf.headers...)
	// If ErrorOnFailure is configured, then errors are returned,
	// but otherwise the message is returned as-is.
	if err != nil {
		return nil, fmt.Errorf("transform: enrich_http_post: %v", err)
	}

	parsed, err := enrichHTTPParseResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("transform: enrich_http_post: %v", err)
	}

	// If SetKey exists, then the response body is written into the message,
	// but otherwise the response is not stored and the message is returned
	// as-is.
	if tf.conf.Object.SetKey != "" {
		if err := msg.SetValue(tf.conf.Object.SetKey, parsed); err != nil {
			return nil, fmt.Errorf("transform: enrich_http_post: %v", err)
		}
	}

	return []*message.Message{msg}, nil
}

func (tf *enrichHTTPPost) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}

func (*enrichHTTPPost) Close(context.Context) error {
	return nil
}
