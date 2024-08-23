package transform

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/brexhq/substation/v2/config"
	iconfig "github.com/brexhq/substation/v2/internal/config"
	"github.com/brexhq/substation/v2/internal/errors"
	"github.com/brexhq/substation/v2/message"
)

type metaErrConfig struct {
	// Transforms that are applied in series with error handling.
	Transforms []config.Config `json:"transforms"`
	// ErrorMessages are regular expressions that match error messages and determine
	// if the error should be caught.
	ErrorMessages []string `json:"error_messages"`

	ID string `json:"id"`
}

func (c *metaErrConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *metaErrConfig) Validate() error {
	for _, t := range c.Transforms {
		if t.Type == "" {
			return fmt.Errorf("transform: %v", errors.ErrMissingRequiredOption)
		}
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

	tfs           []Transformer
	errorMessages []*regexp.Regexp
}

func (tf *metaErr) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	msgs, err := Apply(ctx, tf.tfs, msg)
	if err != nil {
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
