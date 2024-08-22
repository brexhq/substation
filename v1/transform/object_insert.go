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

type objectInsertConfig struct {
	// Value inserted into the object.
	Value interface{} `json:"value"`

	ID     string         `json:"id"`
	Object iconfig.Object `json:"object"`
}

func (c *objectInsertConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *objectInsertConfig) Validate() error {
	if c.Object.TargetKey == "" {
		return fmt.Errorf("object_target_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.Value == nil {
		return fmt.Errorf("value: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

func newObjectInsert(_ context.Context, cfg config.Config) (*objectInsert, error) {
	conf := objectInsertConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform object_insert: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "object_insert"
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	tf := objectInsert{
		conf: conf,
	}

	return &tf, nil
}

type objectInsert struct {
	conf objectInsertConfig
}

func (tf *objectInsert) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	if err := msg.SetValue(tf.conf.Object.TargetKey, tf.conf.Value); err != nil {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
	}

	return []*message.Message{msg}, nil
}

func (tf *objectInsert) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
