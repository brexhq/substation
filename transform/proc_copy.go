package transform

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	_config "github.com/brexhq/substation/internal/config"
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
	if err := _config.Decode(cfg.Settings, &conf); err != nil {
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

func (proc *procCopy) Transform(ctx context.Context, message *mess.Message) ([]*mess.Message, error) {
	// Skip control messages.
	if message.IsControl() {
		return []*mess.Message{message}, nil
	}

	if proc.isObject {
		if err := message.Set(proc.conf.SetKey, message.Get(proc.conf.Key)); err != nil {
			return nil, fmt.Errorf("transform: proc_copy: %v", err)
		}

		return []*mess.Message{message}, nil
	}

	if proc.isFrom {
		res := message.Get(proc.conf.Key).String()

		msg, err := mess.New(
			mess.SetData([]byte(res)),
			mess.SetMetadata(message.Metadata()),
		)
		if err != nil {
			return nil, fmt.Errorf("transform: proc_copy: %v", err)
		}

		return []*mess.Message{msg}, nil
	}

	if proc.isTo {
		msg, err := mess.New(
			mess.SetMetadata(message.Metadata()),
		)
		if err != nil {
			return nil, fmt.Errorf("transform: proc_copy: %v", err)
		}

		if err := msg.Set(proc.conf.SetKey, message.Data()); err != nil {
			return nil, fmt.Errorf("transform: proc_copy: %v", err)
		}

		return []*mess.Message{msg}, nil
	}

	return nil, nil
}
