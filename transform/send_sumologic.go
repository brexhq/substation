package transform

import (
	"bytes"
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
	"github.com/brexhq/substation/message"
)

type sendSumologicConfig struct {
	Buffer iconfig.Buffer `json:"buffer"`

	// URL is the Sumo Logic HTTPS endpoint that objects are sent to.
	URL string `json:"url"`
	// Category is the Sumo Logic source category that overrides the
	// configuration for the HTTPS endpoint.
	//
	// This is required if CategoryKey is not used.
	Category string `json:"category"`
	// CategoryKey retrieves a value from an object that is used as
	// the Sumo Logic source category that overrides the configuration
	// for the HTTPS endpoint. If used, then this overrides Category.
	//
	// This is required if Category is not used.
	CategoryKey string `json:"category_key"`
}

func (c *sendSumologicConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *sendSumologicConfig) Validate() error {
	if c.URL == "" {
		return fmt.Errorf("url: %v", errors.ErrMissingRequiredOption)
	}

	if c.Category == "" && c.CategoryKey == "" {
		return fmt.Errorf("category: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

func newSendSumologic(_ context.Context, cfg config.Config) (*sendSumologic, error) {
	conf := sendSumologicConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: send_sumologic: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: send_sumologic: %v", err)
	}

	tf := sendSumologic{
		conf: conf,
	}

	tf.client.Setup()
	if _, ok := os.LookupEnv("AWS_XRAY_DAEMON_ADDRESS"); ok {
		tf.client.EnableXRay()
	}

	tf.headers = []http.Header{
		{
			Key:   "Content-Type",
			Value: "application/json",
		},
	}

	buffer, err := aggregate.New(aggregate.Config{
		// Sumo Logic limits batches to 1MB.
		Size:     1024 * 1024,
		Count:    conf.Buffer.Count,
		Duration: conf.Buffer.Duration,
	})
	if err != nil {
		return nil, fmt.Errorf("transform: send_aws_s3: %v", err)
	}
	tf.buffer = buffer

	return &tf, nil
}

type sendSumologic struct {
	conf sendSumologicConfig

	// client is safe for concurrent use.
	client  http.HTTP
	headers []http.Header

	mu     sync.Mutex
	buffer *aggregate.Aggregate
}

func (tf *sendSumologic) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	tf.mu.Lock()
	defer tf.mu.Unlock()

	if msg.IsControl() {
		for category := range tf.buffer.GetAll() {
			count := tf.buffer.Count(category)
			if count == 0 {
				continue
			}

			if err := tf.sendPayload(ctx, category); err != nil {
				return nil, fmt.Errorf("transform: send_sumologic: %v", err)
			}
		}

		tf.buffer.ResetAll()
		return []*message.Message{msg}, nil
	}

	if !json.Valid(msg.Data()) {
		return nil, fmt.Errorf("transform: send_sumologic: %v", errMsgInvalidObject)
	}

	category := tf.conf.Category
	if tf.conf.CategoryKey != "" {
		category = msg.GetValue(tf.conf.CategoryKey).String()
	}

	// Sends data to SumoLogic only when the buffer is full.
	if ok := tf.buffer.Add(category, msg.Data()); ok {
		return []*message.Message{msg}, nil
	}

	if err := tf.sendPayload(ctx, category); err != nil {
		return nil, fmt.Errorf("transform: send_sumologic: %v", err)
	}

	// Reset the buffer and add the msg data.
	tf.buffer.Reset(category)
	_ = tf.buffer.Add(category, msg.Data())

	return []*message.Message{msg}, nil
}

func (t *sendSumologic) sendPayload(ctx context.Context, category string) error {
	if t.buffer.Count(category) == 0 {
		return nil
	}

	h := t.headers
	h = append(h, http.Header{
		Key:   "X-Sumo-Category",
		Value: category,
	})

	var buf bytes.Buffer
	for _, i := range t.buffer.Get(category) {
		buf.WriteString(fmt.Sprintf("%s\n", i))
	}

	resp, err := t.client.Post(ctx, t.conf.URL, buf.Bytes(), h...)
	if err != nil {
		// Post errors return metadata.
		return err
	}

	//nolint:errcheck // Response body is discarded to avoid resource leaks.
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()

	return nil
}

func (tf *sendSumologic) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
