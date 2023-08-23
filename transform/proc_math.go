package transform

import (
	"context"
	gojson "encoding/json"
	"fmt"

	"golang.org/x/exp/slices"

	"github.com/brexhq/substation/config"
	_config "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	mess "github.com/brexhq/substation/message"
)

type procMathConfig struct {
	// Key retrieves a value from an object for processing.
	//
	// This is optional for transforms that support processing non-object data.
	Key string `json:"key"`
	// SetKey inserts a processed value into an object.
	//
	// This is optional for transforms that support processing non-object data.
	SetKey string `json:"set_key"`
	// Operation determines the operator applied to the data.
	//
	// Must be one of:
	//
	// - add
	//
	// - subtract
	//
	// - multiply
	//
	// - divide
	Operation string `json:"operation"`
}

type procMath struct {
	conf procMathConfig
}

func newProcMath(_ context.Context, cfg config.Config) (*procMath, error) {
	conf := procMathConfig{}
	if err := _config.Decode(cfg.Settings, &conf); err != nil {
		return nil, err
	}

	// Validate required options.
	if conf.Key == "" || conf.SetKey == "" {
		return nil, fmt.Errorf("transform: proc_math: key %s set_key %s: %v", conf.Key, conf.SetKey, errInvalidDataPattern)
	}

	if conf.Operation == "" {
		return nil, fmt.Errorf("transform: proc_math: operation: %v", errors.ErrMissingRequiredOption)
	}

	if !slices.Contains(
		[]string{
			"add",
			"subtract",
			"multiply",
			"divide",
		},
		conf.Operation) {
		return nil, fmt.Errorf("transform: proc_math: operation %q: %v", conf.Operation, errors.ErrInvalidOption)
	}

	proc := procMath{
		conf: conf,
	}

	return &proc, nil
}

func (proc *procMath) String() string {
	b, _ := gojson.Marshal(proc.conf)
	return string(b)
}

func (*procMath) Close(context.Context) error {
	return nil
}

func (proc *procMath) Transform(ctx context.Context, message *mess.Message) ([]*mess.Message, error) {
	// Skip control messages.
	if message.IsControl() {
		return []*mess.Message{message}, nil
	}

	var value float64
	result := message.Get(proc.conf.Key)
	for i, res := range result.Array() {
		if i == 0 {
			value = res.Float()
			continue
		}

		switch proc.conf.Operation {
		case "add":
			value += res.Float()
		case "subtract":
			value -= res.Float()
		case "multiply":
			value *= res.Float()
		case "divide":
			value /= res.Float()
		}
	}

	if err := message.Set(proc.conf.SetKey, value); err != nil {
		return nil, fmt.Errorf("transform: proc_math: %v", err)
	}

	return []*mess.Message{message}, nil
}
