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

type objectToUintConfig struct {
	Object iconfig.Object `json:"object"`
}

func (c *objectToUintConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *objectToUintConfig) Validate() error {
	if c.Object.Key == "" && c.Object.SetKey != "" {
		return fmt.Errorf("object_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.Object.Key != "" && c.Object.SetKey == "" {
		return fmt.Errorf("object_set_key: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

func newObjectToUint(_ context.Context, cfg config.Config) (*objectToUint, error) {
	conf := objectToUintConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: new_object_to_uint: %v", err)
	}

	tf := objectToUint{
		conf: conf,
	}

	return &tf, nil
}

type objectToUint struct {
	conf objectToUintConfig
}

func (tf *objectToUint) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	value := msg.GetValue(tf.conf.Object.Key)
	if err := msg.SetValue(tf.conf.Object.SetKey, value.Uint()); err != nil {
		return nil, fmt.Errorf("transform: object_to_uint: %v", err)
	}

	return []*message.Message{msg}, nil
}

func (tf *objectToUint) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
