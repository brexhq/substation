package transform

import (
	"context"
	gojson "encoding/json"
	"fmt"
	"strings"

	"github.com/brexhq/substation/config"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/message"
)

type modSplitConfig struct {
	Object configObject `json:"object"`

	// Separator is the string to split the data on.
	Separator string `json:"separator"`
}

type modSplit struct {
	conf modSplitConfig
}

func newModSplit(_ context.Context, cfg config.Config) (*modSplit, error) {
	conf := modSplitConfig{}
	if err := iconfig.Decode(cfg.Settings, &conf); err != nil {
		return nil, fmt.Errorf("transform: new_mod_split: %v", err)
	}

	// Validate required options.
	if conf.Object.Key == "" && conf.Object.SetKey != "" {
		return nil, fmt.Errorf("transform: new_mod_split: object_key: %v", errors.ErrMissingRequiredOption)
	}

	if conf.Object.Key != "" && conf.Object.SetKey == "" {
		return nil, fmt.Errorf("transform: new_mod_split: object_set_key: %v", errors.ErrMissingRequiredOption)
	}

	if conf.Separator == "" {
		return nil, fmt.Errorf("transform: new_mod_split: separator: %v", errors.ErrMissingRequiredOption)
	}

	tf := modSplit{
		conf: conf,
	}

	return &tf, nil
}

func (tf *modSplit) String() string {
	b, _ := gojson.Marshal(tf.conf)
	return string(b)
}

func (*modSplit) Close(context.Context) error {
	return nil
}

func (tf *modSplit) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	// Skip interrupt messages.
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	res := msg.GetObject(tf.conf.Object.Key).String()
	v := strings.Split(res, tf.conf.Separator)

	if err := msg.SetObject(tf.conf.Object.SetKey, v); err != nil {
		return nil, fmt.Errorf("transform: mod_split: %v", err)
	}

	return []*message.Message{msg}, nil
}
