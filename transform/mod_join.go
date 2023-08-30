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

type modJoinConfig struct {
	Object configObject `json:"object"`

	// Separator is the string that joins data from the array.
	Separator string `json:"separator"`
}

type modJoin struct {
	conf modJoinConfig
}

func newModJoin(_ context.Context, cfg config.Config) (*modJoin, error) {
	conf := modJoinConfig{}
	if err := iconfig.Decode(cfg.Settings, &conf); err != nil {
		return nil, fmt.Errorf("transform: new_mod_join: %v", err)
	}

	// Validate required options.
	if conf.Object.Key == "" {
		return nil, fmt.Errorf("transform: new_mod_join: object_key: %v", errors.ErrMissingRequiredOption)
	}

	if conf.Object.SetKey == "" {
		return nil, fmt.Errorf("transform: new_mod_join: object_set_key: %v", errors.ErrMissingRequiredOption)
	}

	if conf.Separator == "" {
		return nil, fmt.Errorf("transform: new_mod_join: separator: %v", errors.ErrMissingRequiredOption)
	}

	tf := modJoin{
		conf: conf,
	}

	return &tf, nil
}

func (tf *modJoin) String() string {
	b, _ := gojson.Marshal(tf.conf)
	return string(b)
}

func (*modJoin) Close(context.Context) error {
	return nil
}

func (tf *modJoin) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	// Skip interrupt messages.
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	// Data is processed by retrieving and iterating the
	// array (Get) containing string values and joining
	// each one with the separator string.
	//
	// Get value:
	// 	{"join":["foo","bar","baz"]}
	// Set value:
	// 	{"join:"foo.bar.baz"}
	var value string
	result := msg.GetObject(tf.conf.Object.Key)
	for i, res := range result.Array() {
		value += res.String()
		if i != len(result.Array())-1 {
			value += tf.conf.Separator
		}
	}

	if err := msg.SetObject(tf.conf.Object.SetKey, value); err != nil {
		return nil, fmt.Errorf("transform: mod_join: %v", err)
	}

	return []*message.Message{msg}, nil
}
