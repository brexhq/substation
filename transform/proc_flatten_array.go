package transform

import (
	"context"
	gojson "encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	_config "github.com/brexhq/substation/internal/config"
	mess "github.com/brexhq/substation/message"
)

type procFlattenArrayConfig struct {
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

type procFlattenArray struct {
	conf procFlattenArrayConfig
}

func newProcFlattenArray(_ context.Context, cfg config.Config) (*procFlattenArray, error) {
	conf := procFlattenArrayConfig{}
	if err := _config.Decode(cfg.Settings, &conf); err != nil {
		return nil, err
	}

	// Validate required options.
	if conf.Key == "" || conf.SetKey == "" {
		return nil, fmt.Errorf("transform: proc_flatten: key %s set_key %s: %v", conf.Key, conf.SetKey, errInvalidDataPattern)
	}

	proc := procFlattenArray{
		conf: conf,
	}

	return &proc, nil
}

func (t *procFlattenArray) String() string {
	b, _ := gojson.Marshal(t.conf)
	return string(b)
}

func (*procFlattenArray) Close(context.Context) error {
	return nil
}

func (t *procFlattenArray) Transform(ctx context.Context, messages ...*mess.Message) ([]*mess.Message, error) {
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
