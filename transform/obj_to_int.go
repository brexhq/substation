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

type objToIntConfig struct {
	Object iconfig.Object `json:"object"`
}

func (c *objToIntConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *objToIntConfig) Validate() error {
	if c.Object.Key == "" && c.Object.SetKey != "" {
		return fmt.Errorf("object_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.Object.Key != "" && c.Object.SetKey == "" {
		return fmt.Errorf("object_set_key: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

type objToInt struct {
	conf objToIntConfig
}

func newObjToInt(_ context.Context, cfg config.Config) (*objToInt, error) {
	conf := objToIntConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: new_object_to_int: %v", err)
	}

	tf := objToInt{
		conf: conf,
	}

	return &tf, nil
}

func (tf *objToInt) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	// Skip interrupt messages.
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	value := msg.GetValue(tf.conf.Object.Key)
	if err := msg.SetValue(tf.conf.Object.SetKey, value.Int()); err != nil {
		return nil, fmt.Errorf("transform: object_to_int: %v", err)
	}

	return []*message.Message{msg}, nil
}

func (tf *objToInt) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}

func (*objToInt) Close(context.Context) error {
	return nil
}
