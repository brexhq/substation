package transform

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/aggregate"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/http"
	"github.com/brexhq/substation/internal/secrets"
	"github.com/brexhq/substation/message"
)

type sendHTTPPostConfig struct {
	// URL is the HTTP(S) endpoint that data is sent to.
	URL string `json:"url"`
	// Headers are an array of objects that contain HTTP headers sent in the request.
	//
	// This is optional and has no default.
	Headers map[string]string `json:"headers"`
	// AuxTransforms are applied to batched data before it is sent.
	AuxTransforms []config.Config `json:"auxiliary_transforms"`

	ID     string         `json:"id"`
	Object iconfig.Object `json:"object"`
	Batch  iconfig.Batch  `json:"batch"`
}

func (c *sendHTTPPostConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *sendHTTPPostConfig) Validate() error {
	if c.URL == "" {
		return fmt.Errorf("url: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

func newSendHTTPPost(_ context.Context, cfg config.Config) (*sendHTTPPost, error) {
	conf := sendHTTPPostConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform send_http_post: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "send_http_post"
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	tf := sendHTTPPost{
		conf: conf,
	}

	tf.client.Setup()
	if _, ok := os.LookupEnv("AWS_XRAY_DAEMON_ADDRESS"); ok {
		tf.client.EnableXRay()
	}

	agg, err := aggregate.New(aggregate.Config{
		Count:    conf.Batch.Count,
		Size:     conf.Batch.Size,
		Duration: conf.Batch.Duration,
	})
	if err != nil {
		return nil, err
	}
	tf.agg = agg

	if len(conf.AuxTransforms) > 0 {
		tf.tforms = make([]Transformer, len(conf.AuxTransforms))
		for i, c := range conf.AuxTransforms {
			t, err := New(context.Background(), c)
			if err != nil {
				return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
			}

			tf.tforms[i] = t
		}
	}

	return &tf, nil
}

type sendHTTPPost struct {
	conf sendHTTPPostConfig

	// client is safe for concurrent use.
	client http.HTTP

	mu     sync.Mutex
	agg    *aggregate.Aggregate
	tforms []Transformer
}

func (tf *sendHTTPPost) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	tf.mu.Lock()
	defer tf.mu.Unlock()

	if msg.IsControl() {
		for key := range tf.agg.GetAll() {
			if tf.agg.Count(key) == 0 {
				continue
			}

			if err := tf.send(ctx, key); err != nil {
				return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
			}
		}

		tf.agg.ResetAll()
		return []*message.Message{msg}, nil
	}

	// If this value does not exist, then all data is batched together.
	key := msg.GetValue(tf.conf.Object.BatchKey).String()
	if ok := tf.agg.Add(key, msg.Data()); ok {
		return []*message.Message{msg}, nil
	}

	if err := tf.send(ctx, key); err != nil {
		return nil, fmt.Errorf("transform %s: %v", err)
	}

	// If data cannot be added after reset, then the batch is misconfgured.
	tf.agg.Reset(key)
	if ok := tf.agg.Add(key, msg.Data()); !ok {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, errSendBatchMisconfigured)
	}

	return []*message.Message{msg}, nil
}

func (tf *sendHTTPPost) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}

func (tf *sendHTTPPost) send(ctx context.Context, key string) error {
	var headers []http.Header
	for k, v := range tf.conf.Headers {
		// Retrieve secret and interpolate with header value.
		v, err := secrets.Interpolate(ctx, v)
		if err != nil {
			return err
		}

		headers = append(headers, http.Header{
			Key:   k,
			Value: v,
		})
	}

	// Retrieve secret and interpolate with URL.
	url, err := secrets.Interpolate(ctx, tf.conf.URL)
	if err != nil {
		return err
	}

	data, err := withTransforms(ctx, tf.tforms, tf.agg.Get(key))
	if err != nil {
		return err
	}

	for _, d := range data {
		resp, err := tf.client.Post(ctx, url, d, headers...)
		if err != nil {
			return err
		}

		//nolint:errcheck // Response body is discarded to avoid resource leaks.
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}

	return nil
}
