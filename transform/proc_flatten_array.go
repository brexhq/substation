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

func (proc *procFlattenArray) String() string {
	b, _ := gojson.Marshal(proc.conf)
	return string(b)
}

func (*procFlattenArray) Close(context.Context) error {
	return nil
}

func (proc *procFlattenArray) Transform(ctx context.Context, message *mess.Message) ([]*mess.Message, error) {
	// Skip control messages.
	if message.IsControl() {
		return []*mess.Message{message}, nil
	}

	var value interface{}
	if proc.conf.Deep {
		value = message.Get(proc.conf.Key + `|@flatten:{"deep":true}`)
	} else {
		value = message.Get(proc.conf.Key + `|@flatten`)
	}

	if err := message.Set(proc.conf.SetKey, value); err != nil {
		return nil, fmt.Errorf("transform: proc_flatten: %v", err)
	}

	return []*mess.Message{message}, nil
}
