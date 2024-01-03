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

type objectToUnsignedIntegerConfig struct {
	Object iconfig.Object `json:"object"`
}

func (c *objectToUnsignedIntegerConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *objectToUnsignedIntegerConfig) Validate() error {
	if c.Object.SourceKey == "" && c.Object.TargetKey != "" {
		return fmt.Errorf("object_source_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.Object.SourceKey != "" && c.Object.TargetKey == "" {
		return fmt.Errorf("object_target_key: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

func newObjectToUnsignedInteger(_ context.Context, cfg config.Config) (*objectToUnsignedInteger, error) {
	conf := objectToUnsignedIntegerConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: object_to_unsigned_integer: %v", err)
	}

	tf := objectToUnsignedInteger{
		conf: conf,
	}

	return &tf, nil
}

type objectToUnsignedInteger struct {
	conf objectToUnsignedIntegerConfig
}

func (tf *objectToUnsignedInteger) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	value := msg.GetValue(tf.conf.Object.SourceKey)
	if !value.Exists() {
		return []*message.Message{msg}, nil
	}

	if err := msg.SetValue(tf.conf.Object.TargetKey, value.Uint()); err != nil {
		return nil, fmt.Errorf("transform: object_to_unsigned_integer: %v", err)
	}

	return []*message.Message{msg}, nil
}

func (tf *objectToUnsignedInteger) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
