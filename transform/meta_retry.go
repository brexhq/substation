package transform

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/message"
)

type metaRetryConfig struct {
	// Transforms that are applied in series, then checked for success
	// based on the condition or errors.
	Transforms []config.Config `json:"transforms"`
	// Condition that must be true for the transforms to be considered
	// a success.
	Condition condition.Config `json:"condition"`

	Retry iconfig.Retry `json:"retry"`
	ID    string        `json:"id"`
}

func (c *metaRetryConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *metaRetryConfig) Validate() error {
	if c.Condition.Operator == "" {
		return fmt.Errorf("condition: %v", errors.ErrMissingRequiredOption)
	}

	for _, t := range c.Transforms {
		if t.Type == "" {
			return fmt.Errorf("transform: %v", errors.ErrMissingRequiredOption)
		}
	}

	return nil
}

func newMetaRetry(ctx context.Context, cfg config.Config) (*metaRetry, error) {
	conf := metaRetryConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform meta_retry: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "meta_retry"
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	tforms := make([]Transformer, len(conf.Transforms))
	for i, t := range conf.Transforms {
		tfer, err := New(ctx, t)
		if err != nil {
			return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
		}

		tforms[i] = tfer
	}

	cnd, err := condition.New(ctx, conf.Condition)
	if err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	dur, err := time.ParseDuration(conf.Retry.Duration)
	if err != nil {
		return nil, fmt.Errorf("transform %s: duration: %v", conf.ID, err)
	}

	tf := metaRetry{
		conf:       conf,
		transforms: tforms,
		condition:  cnd,
		duration:   dur,
	}

	return &tf, nil
}

type metaRetry struct {
	conf metaRetryConfig

	condition  condition.Operator
	transforms []Transformer
	duration   time.Duration
}

func (tf *metaRetry) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	// This handles both data and ctrl messages.
	return tf.retryTransform(ctx, msg, 0)
}

func (tf *metaRetry) retryTransform(ctx context.Context, msg *message.Message, count int) ([]*message.Message, error) {
	if count > tf.conf.Retry.Count {
		return nil, fmt.Errorf("transform %s: limit exceeded", tf.conf.ID)
	}

	// This implements exponential backoff.
	for i := 0; i < count; i++ {
		time.Sleep(tf.duration)
	}

	msgs, err := Apply(ctx, tf.transforms, msg)
	if err != nil {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
	}

	// Every message must pass the condition or else it gets retried.
	for _, m := range msgs {
		if m.IsControl() {
			continue
		}

		ok, err := tf.condition.Operate(ctx, m)
		if err != nil {
			return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
		}

		if !ok {
			return tf.retryTransform(ctx, msg, count+1)
		}
	}

	return msgs, nil
}

func (tf *metaRetry) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
