package transform

import (
	"context"
	gojson "encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/message"
)

type objInsertConfig struct {
	Object iconfig.Object `json:"object"`

	// Value inserted into the object.
	Value interface{} `json:"value"`
}

func (c *objInsertConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *objInsertConfig) Validate() error {
	if c.Object.SetKey == "" {
		return fmt.Errorf("object_set_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.Value == nil {
		return fmt.Errorf("value: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

type objInsert struct {
	conf objInsertConfig
}

func newObjInsert(_ context.Context, cfg config.Config) (*objInsert, error) {
	conf := objInsertConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: new_object_insert: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: new_object_insert: %v", err)
	}

	tf := objInsert{
		conf: conf,
	}

	return &tf, nil
}

func (tf *objInsert) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	// Skip interrupt messages.
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	if err := msg.SetValue(tf.conf.Object.SetKey, tf.conf.Value); err != nil {
		return nil, fmt.Errorf("transform: object_insert: %v", err)
	}

	return []*message.Message{msg}, nil
}

func (tf *objInsert) String() string {
	b, _ := gojson.Marshal(tf.conf)
	return string(b)
}

func (*objInsert) Close(context.Context) error {
	return nil
}
