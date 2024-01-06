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

type objectDeleteConfig struct {
	Object iconfig.Object `json:"object"`
}

func (c *objectDeleteConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *objectDeleteConfig) Validate() error {
	if c.Object.SourceKey == "" {
		return fmt.Errorf("object_source_key: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

func newObjectDelete(_ context.Context, cfg config.Config) (*objectDelete, error) {
	conf := objectDeleteConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: object_delete: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: object_delete: %v", err)
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
		return nil, fmt.Errorf("transform: object_delete: %v", err)
	}

	return []*message.Message{msg}, nil
}

func (tf *objectDelete) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
