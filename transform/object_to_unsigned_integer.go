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
	if c.Object.Key == "" && c.Object.SetKey != "" {
		return fmt.Errorf("object_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.Object.Key != "" && c.Object.SetKey == "" {
		return fmt.Errorf("object_set_key: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

func newObjectToUnsignedInteger(_ context.Context, cfg config.Config) (*objectToUnsignedInteger, error) {
	conf := objectToUnsignedIntegerConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: new_object_to_uint: %v", err)
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

	value := msg.GetValue(tf.conf.Object.Key)
	if err := msg.SetValue(tf.conf.Object.SetKey, value.Uint()); err != nil {
		return nil, fmt.Errorf("transform: object_to_uint: %v", err)
	}

	return []*message.Message{msg}, nil
}

func (tf *objectToUnsignedInteger) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}