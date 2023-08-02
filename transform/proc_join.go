package transform

import (
	"context"
	gojson "encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
	mess "github.com/brexhq/substation/message"
)

type procJoinConfig struct {
	// Key retrieves a value from an object for processing.
	//
	// This is optional for transforms that support processing non-object data.
	Key string `json:"key"`
	// SetKey inserts a processed value into an object.
	//
	// This is optional for transforms that support processing non-object data.
	SetKey string `json:"set_key"`
	// Separator is the string that joins data from the array.
	Separator string `json:"separator"`
}

type procJoin struct {
	conf procJoinConfig
}

func newProcJoin(_ context.Context, cfg config.Config) (*procJoin, error) {
	conf := procJoinConfig{}
	if err := config.Decode(cfg.Settings, &conf); err != nil {
		return nil, err
	}

	// Validate required options.
	if conf.Key == "" || conf.SetKey == "" {
		return nil, fmt.Errorf("transform: proc_join: key %s set_key %s: %v", conf.Key, conf.SetKey, errInvalidDataPattern)
	}

	if conf.Separator == "" {
		return nil, fmt.Errorf("transform: proc_join: separator: %v", errors.ErrMissingRequiredOption)
	}

	proc := procJoin{
		conf: conf,
	}

	return &proc, nil
}

func (t *procJoin) String() string {
	b, _ := gojson.Marshal(t.conf)
	return string(b)
}

func (*procJoin) Close(context.Context) error {
	return nil
}

func (t *procJoin) Transform(ctx context.Context, messages ...*mess.Message) ([]*mess.Message, error) {
	var output []*mess.Message

	for _, message := range messages {
		// Skip control messages.
		if message.IsControl() {
			output = append(output, message)
			continue
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
		result := message.Get(t.conf.Key)
		for i, res := range result.Array() {
			value += res.String()
			if i != len(result.Array())-1 {
				value += t.conf.Separator
			}
		}

		if err := message.Set(t.conf.SetKey, value); err != nil {
			return nil, fmt.Errorf("transform: proc_join: %v", err)
		}

		output = append(output, message)
	}

	return output, nil
}
