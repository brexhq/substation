package transform

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"github.com/brexhq/substation/v2/condition"
	"github.com/brexhq/substation/v2/config"
	iconfig "github.com/brexhq/substation/v2/internal/config"
	"github.com/brexhq/substation/v2/internal/errors"
	"github.com/brexhq/substation/v2/message"
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
	Condition config.Config `json:"condition"`

	Retry iconfig.Retry `json:"retry"`
	ID    string        `json:"id"`
}

func (c *metaRetryConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *metaRetryConfig) Validate() error {
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

	errs := make([]*regexp.Regexp, len(conf.Retry.ErrorMessages))
	for i, e := range conf.Retry.ErrorMessages {
		r, err := regexp.Compile(e)
		if err != nil {
			return nil, fmt.Errorf("transform %s: error_messages: %v", conf.ID, err)
		}

		errs[i] = r
	}

	tf := metaRetry{
		conf:          conf,
		transforms:    tforms,
		condition:     cnd,
		delay:         del,
		errorMessages: errs,
	}

	return &tf, nil
}

type metaRetry struct {
	conf metaRetryConfig

	condition     condition.Inspector
	transforms    []Transformer
	delay         time.Duration
	errorMessages []*regexp.Regexp
}

func (tf *metaRetry) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	var counter int

LOOP:
	for {
		// If the retry count is set to 0, then this will retry forever.
		if tf.conf.Retry.Count > 0 && counter > tf.conf.Retry.Count {
			break
		}

		// Implements constant backoff. The first iteration is skipped.
		if counter > 0 {
			time.Sleep(tf.delay)
		}

		counter++

		// This must operate on a copy of the message to avoid
		// modifying the original message in case the transform
		// fails.
		cMsg := *msg
		msgs, err := Apply(ctx, tf.transforms, &cMsg)
		if err != nil {
			for _, r := range tf.errorMessages {
				if r.MatchString(err.Error()) {
					continue LOOP
				}
			}

			return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
		}

		for _, m := range msgs {
			if m.IsControl() {
				continue
			}

			ok, err := tf.condition.Inspect(ctx, m)
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
