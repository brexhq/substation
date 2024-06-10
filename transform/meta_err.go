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
	Transform config.Config `json:"transform"`
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
	if c.Transform.Type == "" {
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

	tf, err := New(ctx, conf.Transform)
	if err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	errMsgs := make([]*regexp.Regexp, len(conf.ErrorMessages))
	for i, eMsg := range conf.ErrorMessages {
		r, err := regexp.Compile(eMsg)
		if err != nil {
			return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
		}

		errMsgs[i] = r
	}

	meta := metaErr{
		conf:          conf,
		tf:            tf,
		errorMessages: errMsgs,
	}

	return &meta, nil
}

type metaErr struct {
	conf metaErrConfig

	tf            Transformer
	errorMessages []*regexp.Regexp
}

func (tf *metaErr) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	msgs, err := tf.tf.Transform(ctx, msg)
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
