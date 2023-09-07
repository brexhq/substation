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

type objToFloatConfig struct {
	Object iconfig.Object `json:"object"`
}

func (c *objToFloatConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *objToFloatConfig) Validate() error {
	if c.Object.Key == "" && c.Object.SetKey != "" {
		return fmt.Errorf("object_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.Object.Key != "" && c.Object.SetKey == "" {
		return fmt.Errorf("object_set_key: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

type objToFloat struct {
	conf objToFloatConfig
}

func newObjToFloat(_ context.Context, cfg config.Config) (*objToFloat, error) {
	conf := objToFloatConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: new_object_to_float: %v", err)
	}

	tf := objToFloat{
		conf: conf,
	}

	return &tf, nil
}

func (tf *objToFloat) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	// Skip interrupt messages.
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	value := msg.GetValue(tf.conf.Object.Key)
	if err := msg.SetValue(tf.conf.Object.SetKey, value.Float()); err != nil {
		return nil, fmt.Errorf("transform: object_to_float: %v", err)
	}

	return []*message.Message{msg}, nil
}

func (tf *objToFloat) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}

func (*objToFloat) Close(context.Context) error {
	return nil
}
