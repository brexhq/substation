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

type objectToFloatConfig struct {
	Object iconfig.Object `json:"object"`
}

func (c *objectToFloatConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *objectToFloatConfig) Validate() error {
	if c.Object.Key == "" && c.Object.SetKey != "" {
		return fmt.Errorf("object_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.Object.Key != "" && c.Object.SetKey == "" {
		return fmt.Errorf("object_set_key: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

func newObjectToFloat(_ context.Context, cfg config.Config) (*objectToFloat, error) {
	conf := objectToFloatConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: object_to_float: %v", err)
	}

	tf := objectToFloat{
		conf: conf,
	}

	return &tf, nil
}

type objectToFloat struct {
	conf objectToFloatConfig
}

func (tf *objectToFloat) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	value := msg.GetValue(tf.conf.Object.Key)
	if !value.Exists() {
		return []*message.Message{msg}, nil
	}
	
	if err := msg.SetValue(tf.conf.Object.SetKey, value.Float()); err != nil {
		return nil, fmt.Errorf("transform: object_to_float: %v", err)
	}

	return []*message.Message{msg}, nil
}

func (tf *objectToFloat) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
