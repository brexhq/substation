package transform

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"

	iconfig "github.com/brexhq/substation/v2/internal/config"
)

type objectToStringConfig struct {
	ID     string         `json:"id"`
	Object iconfig.Object `json:"object"`
}

func (c *objectToStringConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *objectToStringConfig) Validate() error {
	if c.Object.SourceKey == "" && c.Object.TargetKey != "" {
		return fmt.Errorf("object_source_key: %v", iconfig.ErrMissingRequiredOption)
	}

	if c.Object.SourceKey != "" && c.Object.TargetKey == "" {
		return fmt.Errorf("object_target_key: %v", iconfig.ErrMissingRequiredOption)
	}

	return nil
}

func newObjectToString(_ context.Context, cfg config.Config) (*objectToString, error) {
	conf := objectToStringConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform object_to_string: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "object_to_string"
	}

	tf := objectToString{
		conf: conf,
	}

	return &tf, nil
}

type objectToString struct {
	conf objectToStringConfig
}

func (tf *objectToString) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.HasFlag(message.IsControl) {
		return []*message.Message{msg}, nil
	}

	value := msg.GetValue(tf.conf.Object.SourceKey)
	if skipMessage(msg, value) {
		return []*message.Message{msg}, nil
	}

	if err := msg.SetValue(tf.conf.Object.TargetKey, value.String()); err != nil {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
	}

	return []*message.Message{msg}, nil
}

func (tf *objectToString) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
