package transform

import (
	"context"
	gojson "encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	mess "github.com/brexhq/substation/message"
)

type procFlattenConfig struct {
	// Key retrieves a value from an object for processing.
	//
	// This is optional for transforms that support processing non-object data.
	Key string `json:"key"`
	// SetKey inserts a processed value into an object.
	//
	// This is optional for transforms that support processing non-object data.
	SetKey string `json:"set_key"`
	// Deep determines if arrays should be deeply flattened.
	//
	// This is optional and defaults to false.
	Deep bool `json:"deep"`
}

type procFlatten struct {
	conf procFlattenConfig
}

func newProcFlatten(_ context.Context, cfg config.Config) (*procFlatten, error) {
	conf := procFlattenConfig{}
	if err := config.Decode(cfg.Settings, &conf); err != nil {
		return nil, err
	}

	// Validate required options.
	if conf.Key == "" || conf.SetKey == "" {
		return nil, fmt.Errorf("transform: proc_flatten: key %s set_key %s: %v", conf.Key, conf.SetKey, errInvalidDataPattern)
	}

	proc := procFlatten{
		conf: conf,
	}

	return &proc, nil
}

func (t *procFlatten) String() string {
	b, _ := gojson.Marshal(t.conf)
	return string(b)
}

func (*procFlatten) Close(context.Context) error {
	return nil
}

func (t *procFlatten) Transform(ctx context.Context, messages ...*mess.Message) ([]*mess.Message, error) {
	var output []*mess.Message

	for _, message := range messages {
		// Skip control messages.
		if message.IsControl() {
			output = append(output, message)
			continue
		}

		var value interface{}
		if t.conf.Deep {
			value = message.Get(t.conf.Key + `|@flatten:{"deep":true}`)
		} else {
			value = message.Get(t.conf.Key + `|@flatten`)
		}

		if err := message.Set(t.conf.SetKey, value); err != nil {
			return nil, fmt.Errorf("transform: proc_flatten: %v", err)
		}

		output = append(output, message)
	}

	return output, nil
}