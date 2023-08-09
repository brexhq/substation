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

func (t *procInsert) String() string {
	b, _ := gojson.Marshal(t.conf)
	return string(b)
}

func (*procInsert) Close(context.Context) error {
	return nil
}

func (t *procInsert) Transform(ctx context.Context, messages ...*mess.Message) ([]*mess.Message, error) {
	var output []*mess.Message

	for _, message := range messages {
		// Skip control messages.
		if message.IsControl() {
			output = append(output, message)
			continue
		}

		if err := message.Set(t.conf.SetKey, t.conf.Value); err != nil {
			return nil, fmt.Errorf("transform: proc_insert: %v", err)
		}

		output = append(output, message)
	}

	return output, nil
}
