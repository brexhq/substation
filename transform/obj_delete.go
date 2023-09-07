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

type objDeleteConfig struct {
	Object iconfig.Object `json:"object"`
}

func (c *objDeleteConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *objDeleteConfig) Validate() error {
	if c.Object.Key == "" {
		return fmt.Errorf("object_key: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

type objDelete struct {
	conf objDeleteConfig
}

func newObjDelete(_ context.Context, cfg config.Config) (*objDelete, error) {
	conf := objDeleteConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: new_object_delete: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: new_object_delete: %v", err)
	}

	proc := objDelete{
		conf: conf,
	}

	return &proc, nil
}

func (tf *objDelete) Transform(_ context.Context, msg *message.Message) ([]*message.Message, error) {
	// Skip interrupt messages.
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	if err := msg.DeleteValue(tf.conf.Object.Key); err != nil {
		return nil, fmt.Errorf("transform: object_delete: %v", err)
	}

	return []*message.Message{msg}, nil
}

func (tf *objDelete) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}

func (*objDelete) Close(context.Context) error {
	return nil
}
