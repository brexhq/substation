package transform

import (
	"context"
	gojson "encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/message"
)

type modDeleteConfig struct {
	Object configObject `json:"object"`
}

type modDelete struct {
	conf modDeleteConfig
}

func newModDelete(_ context.Context, cfg config.Config) (*modDelete, error) {
	conf := modDeleteConfig{}
	if err := iconfig.Decode(cfg.Settings, &conf); err != nil {
		return nil, fmt.Errorf("transform: new_mod_delete: %v", err)
	}

	proc := modDelete{
		conf: conf,
	}

	return &proc, nil
}

func (tf *modDelete) String() string {
	b, _ := gojson.Marshal(tf.conf)
	return string(b)
}

func (*modDelete) Close(context.Context) error {
	return nil
}

func (tf *modDelete) Transform(_ context.Context, msg *message.Message) ([]*message.Message, error) {
	// Skip interrupt messages.
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	if err := msg.DeleteObject(tf.conf.Object.Key); err != nil {
		return nil, fmt.Errorf("transform: mod_delete: %v", err)
	}

	return []*message.Message{msg}, nil
}
