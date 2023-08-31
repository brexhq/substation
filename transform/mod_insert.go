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

type modInsertConfig struct {
	Object configObject `json:"object"`

	// Value inserted into the object.
	Value interface{} `json:"value"`
}

type modInsert struct {
	conf modInsertConfig
}

func newModInsert(_ context.Context, cfg config.Config) (*modInsert, error) {
	conf := modInsertConfig{}
	if err := iconfig.Decode(cfg.Settings, &conf); err != nil {
		return nil, fmt.Errorf("transform: new_mod_insert: %v", err)
	}

	// Validate required options.
	if conf.Object.SetKey == "" {
		return nil, fmt.Errorf("transform: new_mod_insert: object_set_key: %v", errors.ErrMissingRequiredOption)
	}

	if conf.Value == nil {
		return nil, fmt.Errorf("transform: new_mod_insert: value: %v", errors.ErrMissingRequiredOption)
	}

	tf := modInsert{
		conf: conf,
	}

	return &tf, nil
}

func (tf *modInsert) String() string {
	b, _ := gojson.Marshal(tf.conf)
	return string(b)
}

func (*modInsert) Close(context.Context) error {
	return nil
}

func (tf *modInsert) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	// Skip interrupt messages.
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	if err := msg.SetObject(tf.conf.Object.SetKey, tf.conf.Value); err != nil {
		return nil, fmt.Errorf("transform: mod_insert: %v", err)
	}

	return []*message.Message{msg}, nil
}
