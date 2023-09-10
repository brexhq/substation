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

type objectToBoolConfig struct {
	Object iconfig.Object `json:"object"`
}

func (c *objectToBoolConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *objectToBoolConfig) Validate() error {
	if c.Object.Key == "" && c.Object.SetKey != "" {
		return fmt.Errorf("object_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.Object.Key != "" && c.Object.SetKey == "" {
		return fmt.Errorf("object_set_key: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

func newObjectToBool(_ context.Context, cfg config.Config) (*objectToBool, error) {
	conf := objectToBoolConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: new_object_to_bool: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: new_object_to_bool: %v", err)
	}

	tf := objectToBool{
		conf: conf,
	}

	return &tf, nil
}

type objectToBool struct {
	conf objectToBoolConfig
}

func (tf *objectToBool) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}

func (*objectToBool) Close(context.Context) error {
	return nil
}

func (tf *objectToBool) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	value := msg.GetValue(tf.conf.Object.Key)
	if err := msg.SetValue(tf.conf.Object.SetKey, value.Bool()); err != nil {
		return nil, fmt.Errorf("transform: object_to_bool: %v", err)
	}

	return []*message.Message{msg}, nil
}
