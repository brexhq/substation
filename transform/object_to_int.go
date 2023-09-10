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

type objectToIntConfig struct {
	Object iconfig.Object `json:"object"`
}

func (c *objectToIntConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *objectToIntConfig) Validate() error {
	if c.Object.Key == "" && c.Object.SetKey != "" {
		return fmt.Errorf("object_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.Object.Key != "" && c.Object.SetKey == "" {
		return fmt.Errorf("object_set_key: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

func newObjectToInt(_ context.Context, cfg config.Config) (*objectToInt, error) {
	conf := objectToIntConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: new_object_to_int: %v", err)
	}

	tf := objectToInt{
		conf: conf,
	}

	return &tf, nil
}

type objectToInt struct {
	conf objectToIntConfig
}

func (tf *objectToInt) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	value := msg.GetValue(tf.conf.Object.Key)
	if err := msg.SetValue(tf.conf.Object.SetKey, value.Int()); err != nil {
		return nil, fmt.Errorf("transform: object_to_int: %v", err)
	}

	return []*message.Message{msg}, nil
}

func (tf *objectToInt) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}

func (*objectToInt) Close(context.Context) error {
	return nil
}
