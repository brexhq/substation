package transform

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	mess "github.com/brexhq/substation/message"
)

type procCopyConfig struct {
	// Key retrieves a value from an object for processing.
	//
	// This is optional for transforms that support processing non-object data.
	Key string `json:"key"`
	// SetKey inserts a processed value into an object.
	//
	// This is optional for transforms that support processing non-object data.
	SetKey string `json:"set_key"`
}

type procCopy struct {
	conf procCopyConfig

	isObject bool
	isFrom   bool
	isTo     bool
}

func newProcCopy(_ context.Context, cfg config.Config) (*procCopy, error) {
	conf := procCopyConfig{}
	if err := config.Decode(cfg.Settings, &conf); err != nil {
		return nil, err
	}

	// Validate required options.
	if conf.Key == "" && conf.SetKey == "" {
		return nil, fmt.Errorf("transform: proc_copy: key %s set_key %s: %w", conf.Key, conf.SetKey, errInvalidDataPattern)
	}

	proc := &procCopy{
		conf:     conf,
		isObject: conf.Key != "" && conf.SetKey != "",
		isFrom:   conf.Key != "" && conf.SetKey == "",
		isTo:     conf.Key == "" && conf.SetKey != "",
	}

	return proc, nil
}

func (t *procCopy) String() string {
	b, _ := json.Marshal(t.conf)
	return string(b)
}

func (*procCopy) Close(context.Context) error {
	return nil
}

func (t *procCopy) Transform(ctx context.Context, messages ...*mess.Message) ([]*mess.Message, error) {
	var output []*mess.Message

	for _, message := range messages {
		// Skip control messages.
		if message.IsControl() {
			output = append(output, message)
			continue
		}

		switch {
		case t.isObject:
			if err := message.Set(t.conf.SetKey, message.Get(t.conf.Key)); err != nil {
				return nil, fmt.Errorf("transform: proc_copy: %v", err)
			}

			output = append(output, message)
		case t.isFrom:
			res := message.Get(t.conf.Key).String()

			msg, err := mess.New(
				mess.SetData([]byte(res)),
				mess.SetMetadata(message.Metadata()),
			)
			if err != nil {
				return nil, fmt.Errorf("transform: proc_copy: %v", err)
			}

			output = append(output, msg)
		case t.isTo:
			msg, err := mess.New(
				mess.SetMetadata(message.Metadata()),
			)
			if err != nil {
				return nil, fmt.Errorf("transform: proc_copy: %v", err)
			}

			if err := msg.Set(t.conf.SetKey, message.Data()); err != nil {
				return nil, fmt.Errorf("transform: proc_copy: %v", err)
			}

			output = append(output, msg)
		}
	}

	return output, nil
}
