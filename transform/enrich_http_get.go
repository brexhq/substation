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

type enrichHTTPGetConfig struct {
	// URL is the HTTP(S) endpoint that data is retrieved from.
	//
	// If the substring ${DATA} is in the URL, then the URL is interpolated with
	// data (either the value from Object.SourceKey or the raw data). URLs may be optionally
	// interpolated with secrets (e.g., ${SECRET:FOO}).
	URL string `json:"url"`
	// Headers are an array of objects that contain HTTP headers sent in the request.
	// Values may be optionally interpolated with secrets (e.g., ${SECRET:FOO}).
	//
	// This is optional and has no default.
	Headers map[string]string `json:"headers"`

	ID     string         `json:"id"`
	Object iconfig.Object `json:"object"`
}

func (c *enrichHTTPGetConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *enrichHTTPGetConfig) Validate() error {
	if c.URL == "" {
		return fmt.Errorf("url: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

func newEnrichHTTPGet(ctx context.Context, cfg config.Config) (*enrichHTTPGet, error) {
	conf := enrichHTTPGetConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform enrich_http_get: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "enrich_http_get"
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	tf := enrichHTTPGet{
		conf: conf,
	}

	tf.client.Setup()
	for k, v := range conf.Headers {
		// Retrieve secret and interpolate with header value.
		v, err := secrets.Interpolate(ctx, v)
		if err != nil {
			return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
		}

		tf.headers = append(tf.headers, http.Header{
			Key:   k,
			Value: v,
		})
	}

	return &tf, nil
}

type enrichHTTPGet struct {
	conf enrichHTTPGetConfig

	// client is safe for concurrent use.
	client  http.HTTP
	headers []http.Header
}

func (tf *enrichHTTPGet) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
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
	// The URL is always interpolated with the substring ${DATA}.
	url := tf.conf.URL
	if strings.Contains(url, enrichHTTPInterp) {
		if tf.conf.Object.SourceKey != "" {
			value := msg.GetValue(tf.conf.Object.SourceKey)
			if !value.Exists() {
				return []*message.Message{msg}, nil
			}

			url = strings.ReplaceAll(url, enrichHTTPInterp, value.String())
		} else {
			url = strings.ReplaceAll(url, enrichHTTPInterp, string(msg.Data()))
		}
	}

	// Retrieve secret and interpolate with URL
	url, err := secrets.Interpolate(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
	}

	// resp.Body is closed by enrichHTTPParseResponse.
	resp, err := tf.client.Get(ctx, url, tf.headers...)
	if err != nil {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
	}

	parsed, err := enrichHTTPParseResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("transform	%s: %v", tf.conf.ID, err)
	}

	// If TargetKey is set, then the response body is stored in the message.
	// Otherwise, the response body overwrites the message data.
	if tf.conf.Object.TargetKey != "" {
		if err := msg.SetValue(tf.conf.Object.TargetKey, parsed); err != nil {
			return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
		}

		return []*message.Message{msg}, nil
	}

	msg.SetData(parsed)
	return []*message.Message{msg}, nil
}

func (tf *enrichHTTPGet) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
