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

type objectToBooleanConfig struct {
	Object iconfig.Object `json:"object"`
}

func (c *objectToBooleanConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *objectToBooleanConfig) Validate() error {
	if c.Object.Key == "" && c.Object.SetKey != "" {
		return fmt.Errorf("object_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.Object.Key != "" && c.Object.SetKey == "" {
		return fmt.Errorf("object_set_key: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

func newObjectToBoolean(_ context.Context, cfg config.Config) (*objectToBoolean, error) {
	conf := objectToBooleanConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: object_to_boolean: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: object_to_boolean: %v", err)
	}

	tf := objectToBoolean{
		conf: conf,
	}

	return &tf, nil
}

type objectToBoolean struct {
	conf objectToBooleanConfig
}

func (tf *objectToBoolean) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}

func (tf *objectToBoolean) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	value := msg.GetValue(tf.conf.Object.Key)
	if !value.Exists() {
		return []*message.Message{msg}, nil
	}

	if err := msg.SetValue(tf.conf.Object.SetKey, value.Bool()); err != nil {
		return nil, fmt.Errorf("transform: object_to_boolean: %v", err)
	}

	return []*message.Message{msg}, nil
}
