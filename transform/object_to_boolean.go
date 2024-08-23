package transform

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation/v2/config"
	iconfig "github.com/brexhq/substation/v2/internal/config"
	"github.com/brexhq/substation/v2/internal/errors"
	"github.com/brexhq/substation/v2/message"
)

type objectToBooleanConfig struct {
	ID     string         `json:"id"`
	Object iconfig.Object `json:"object"`
}

func (c *objectToBooleanConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *objectToBooleanConfig) Validate() error {
	if c.Object.SourceKey == "" && c.Object.TargetKey != "" {
		return fmt.Errorf("object_source_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.Object.SourceKey != "" && c.Object.TargetKey == "" {
		return fmt.Errorf("object_target_key: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

func newObjectToBoolean(_ context.Context, cfg config.Config) (*objectToBoolean, error) {
	conf := objectToBooleanConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform object_to_boolean: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "object_to_boolean"
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	tf := objectToBoolean{
		conf: conf,
	}

	return &tf, nil
}

type objectToBoolean struct {
	conf objectToBooleanConfig
}

func (tf *objectToBoolean) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}

func (tf *objectToBoolean) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	value := msg.GetValue(tf.conf.Object.SourceKey)
	if !value.Exists() {
		return []*message.Message{msg}, nil
	}

	if err := msg.SetValue(tf.conf.Object.TargetKey, value.Bool()); err != nil {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
	}

	return []*message.Message{msg}, nil
}
