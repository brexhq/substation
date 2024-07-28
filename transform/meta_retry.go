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

// errMetaRetryLimitReached is returned when the configured retry
// limit is reached. Other transforms may try to catch this error, so
// any update to the variable's value is considered a BREAKING CHANGE.
var errMetaRetryLimitReached = fmt.Errorf("retry limit reached")

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

	del, err := time.ParseDuration(conf.Retry.Delay)
	if err != nil {
		return nil, fmt.Errorf("transform %s: delay: %v", conf.ID, err)
	}

	tf := metaRetry{
		conf:       conf,
		transforms: tforms,
		condition:  cnd,
		delay:      del,
	}

	return &tf, nil
}

type metaRetry struct {
	conf metaRetryConfig

	condition  condition.Operator
	transforms []Transformer
	delay      time.Duration
}

func (tf *metaRetry) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
LOOP:
	// The first iteration is not a retry, so 1 is added to the
	// configured count. If this isn't done, then the retries will
	// be off by one. The first iteration should never sleep.
	for i := 0; i < tf.conf.Retry.Count+1; i++ {
		// Implements constant backoff.
		if i > 0 {
			time.Sleep(tf.delay)
		}

		// This must operate on a copy of the message to avoid
		// modifying the original message in case the transform
		// fails.
		cMsg := *msg
		msgs, err := Apply(ctx, tf.transforms, &cMsg)
		if err != nil {
			return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
		}

		for _, m := range msgs {
			if m.IsControl() {
				continue
			}

			ok, err := tf.condition.Operate(ctx, m)
			if err != nil {
				return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
			}

			// Any condition failure immediately restarts the loop.
			if !ok {
				continue LOOP
			}
		}

		return msgs, nil
	}

	return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, errMetaRetryLimitReached)
}

func (tf *metaRetry) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
