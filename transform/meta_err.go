package transform

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/message"
)

type metaErrConfig struct {
	// Transform that is applied with error handling.
	Transform config.Config `json:"transform"`
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
		return nil, fmt.Errorf("transform: meta_err: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: meta_err: %v", err)
	}

	tf, err := New(ctx, conf.Transform)
	if err != nil {
		return nil, fmt.Errorf("transform: meta_err: %v", err)
	}

	meta := metaErr{
		conf: conf,
		tf:   tf,
	}

	return &meta, nil
}

type metaErr struct {
	conf metaErrConfig

	tf Transformer
}

func (tf *metaErr) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	msgs, err := tf.tf.Transform(ctx, msg)
	if err != nil {
		//nolint: nilerr // ignore non-nil error
		return []*message.Message{msg}, nil
	}

	return msgs, nil
}

func (tf *metaErr) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
