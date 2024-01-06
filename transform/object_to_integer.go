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

type objectToIntegerConfig struct {
	Object iconfig.Object `json:"object"`
}

func (c *objectToIntegerConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *objectToIntegerConfig) Validate() error {
	if c.Object.SourceKey == "" && c.Object.TargetKey != "" {
		return fmt.Errorf("object_source_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.Object.SourceKey != "" && c.Object.TargetKey == "" {
		return fmt.Errorf("object_target_key: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

func newObjectToInteger(_ context.Context, cfg config.Config) (*objectToInteger, error) {
	conf := objectToIntegerConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: object_to_integer: %v", err)
	}

	tf := objectToInteger{
		conf: conf,
	}

	return &tf, nil
}

type objectToInteger struct {
	conf objectToIntegerConfig
}

func (tf *objectToInteger) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	value := msg.GetValue(tf.conf.Object.SourceKey)
	if !value.Exists() {
		return []*message.Message{msg}, nil
	}

	if err := msg.SetValue(tf.conf.Object.TargetKey, value.Int()); err != nil {
		return nil, fmt.Errorf("transform: object_to_integer: %v", err)
	}

	return []*message.Message{msg}, nil
}

func (tf *objectToInteger) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
