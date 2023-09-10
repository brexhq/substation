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

type objectToStrConfig struct {
	Object iconfig.Object `json:"object"`
}

func (c *objectToStrConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *objectToStrConfig) Validate() error {
	if c.Object.Key == "" && c.Object.SetKey != "" {
		return fmt.Errorf("object_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.Object.Key != "" && c.Object.SetKey == "" {
		return fmt.Errorf("object_set_key: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

func newObjectToStr(_ context.Context, cfg config.Config) (*objectToStr, error) {
	conf := objectToStrConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: new_object_to_str: %v", err)
	}

	tf := objectToStr{
		conf: conf,
	}

	return &tf, nil
}

type objectToStr struct {
	conf objectToStrConfig
}

func (tf *objectToStr) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	value := msg.GetValue(tf.conf.Object.Key)
	if err := msg.SetValue(tf.conf.Object.SetKey, value.String()); err != nil {
		return nil, fmt.Errorf("transform: object_to_str: %v", err)
	}

	return []*message.Message{msg}, nil
}

func (tf *objectToStr) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}

func (*objectToStr) Close(context.Context) error {
	return nil
}
