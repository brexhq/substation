package transform

import (
	"context"
	gojson "encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	_config "github.com/brexhq/substation/internal/config"
	mess "github.com/brexhq/substation/message"
)

type procInsertConfig struct {
	// SetKey inserts a processed value into an object.
	//
	// This is optional for transforms that support processing non-object data.
	SetKey string `json:"set_key"`
	// Value inserted into the object.
	Value interface{} `json:"value"`
}

type procInsert struct {
	conf procInsertConfig
}

func newProcInsert(_ context.Context, cfg config.Config) (*procInsert, error) {
	conf := procInsertConfig{}
	if err := _config.Decode(cfg.Settings, &conf); err != nil {
		return nil, err
	}

	// Validate required options.
	if conf.SetKey == "" {
		return nil, fmt.Errorf("transform: proc_insert: missing set_key: %v", errInvalidDataPattern)
	}

	proc := procInsert{
		conf: conf,
	}

	return &proc, nil
}

func (proc *procInsert) String() string {
	b, _ := gojson.Marshal(proc.conf)
	return string(b)
}

func (*procInsert) Close(context.Context) error {
	return nil
}

func (proc *procInsert) Transform(ctx context.Context, message *mess.Message) ([]*mess.Message, error) {
	// Skip control messages.
	if message.IsControl() {
		return []*mess.Message{message}, nil
	}

	if err := message.Set(proc.conf.SetKey, proc.conf.Value); err != nil {
		return nil, fmt.Errorf("transform: proc_insert: %v", err)
	}

	return []*mess.Message{message}, nil
}
