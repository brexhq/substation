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

type objToUintConfig struct {
	Object iconfig.Object `json:"object"`
}

func (c *objToUintConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *objToUintConfig) Validate() error {
	if c.Object.Key == "" && c.Object.SetKey != "" {
		return fmt.Errorf("object_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.Object.Key != "" && c.Object.SetKey == "" {
		return fmt.Errorf("object_set_key: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

type objToUint struct {
	conf objToUintConfig
}

func newObjToUint(_ context.Context, cfg config.Config) (*objToUint, error) {
	conf := objToUintConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: new_object_to_uint: %v", err)
	}

	tf := objToUint{
		conf: conf,
	}

	return &tf, nil
}

func (tf *objToUint) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	// Skip interrupt messages.
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	value := msg.GetValue(tf.conf.Object.Key)
	if err := msg.SetValue(tf.conf.Object.SetKey, value.Uint()); err != nil {
		return nil, fmt.Errorf("transform: object_to_uint: %v", err)
	}

	return []*message.Message{msg}, nil
}

func (tf *objToUint) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}

func (*objToUint) Close(context.Context) error {
	return nil
}
