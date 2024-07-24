package transform

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/brexhq/substation/config"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/message"
)

type metaErrConfig struct {
	// Transform that is applied with error handling.
	//
	// Deprecated: Transform exists for backwards compatibility and will be
	// removed in a future release. Use Transforms instead.
	Transform config.Config `json:"transform"`
	// Transforms that are applied in series with error handling.
	Transforms []config.Config `json:"transforms"`

	// ErrorMessages are regular expressions that match error messages and determine
	// if the error should be caught.
	//
	// This is optional and defaults to an empty list (all errors are caught).
	ErrorMessages []string `json:"error_messages"`

	ID string `json:"id"`
}

func (c *metaErrConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *metaErrConfig) Validate() error {
	if c.Transform.Type == "" && len(c.Transforms) == 0 {
		return fmt.Errorf("transform: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

func newMetaErr(ctx context.Context, cfg config.Config) (*metaErr, error) {
	conf := metaErrConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform meta_err: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "meta_err"
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	tf := metaErr{
		conf: conf,
	}

	if conf.Transform.Type != "" {
		tfer, err := New(ctx, conf.Transform)
		if err != nil {
			return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
		}

		tf.tf = tfer
	}

	tf.tfs = make([]Transformer, len(conf.Transforms))
	for i, t := range conf.Transforms {
		tfer, err := New(ctx, t)
		if err != nil {
			return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
		}

		tf.tfs[i] = tfer
	}

	tf.errorMessages = make([]*regexp.Regexp, len(conf.ErrorMessages))
	for i, eMsg := range conf.ErrorMessages {
		r, err := regexp.Compile(eMsg)
		if err != nil {
			return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
		}

		tf.errorMessages[i] = r
	}

	return &tf, nil
}

type metaErr struct {
	conf metaErrConfig

	tf            Transformer
	tfs           []Transformer
	errorMessages []*regexp.Regexp
}

func (tf *metaErr) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	var msgs []*message.Message
	var err error

	if len(tf.tfs) > 0 {
		msgs, err = Apply(ctx, tf.tfs, msg)
	} else {
		msgs, err = tf.tf.Transform(ctx, msg)
	}

	if err != nil {
		if len(tf.errorMessages) == 0 {
			return []*message.Message{msg}, nil
		}

		for _, e := range tf.errorMessages {
			if e.MatchString(err.Error()) {
				return []*message.Message{msg}, nil
			}
		}

		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
	}

	return msgs, nil
}

func (tf *metaErr) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
