package transform

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"

	iconfig "github.com/brexhq/substation/v2/internal/config"
)

type objectDeleteConfig struct {
	ID     string         `json:"id"`
	Object iconfig.Object `json:"object"`
}

func (c *objectDeleteConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *objectDeleteConfig) Validate() error {
	if c.Object.SourceKey == "" {
		return fmt.Errorf("object_source_key: %v", iconfig.ErrMissingRequiredOption)
	}

	return nil
}

func newObjectDelete(_ context.Context, cfg config.Config) (*objectDelete, error) {
	conf := objectDeleteConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform object_delete: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "object_delete"
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	proc := objectDelete{
		conf: conf,
	}

	return &proc, nil
}

type objectDelete struct {
	conf objectDeleteConfig
}

func (tf *objectDelete) Transform(_ context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	if err := msg.DeleteValue(tf.conf.Object.SourceKey); err != nil {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
	}

	return []*message.Message{msg}, nil
}

func (tf *objectDelete) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
