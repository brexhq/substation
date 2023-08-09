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

func (t *procMath) String() string {
	b, _ := gojson.Marshal(t.conf)
	return string(b)
}

func (*procMath) Close(context.Context) error {
	return nil
}

func (t *procMath) Transform(ctx context.Context, messages ...*mess.Message) ([]*mess.Message, error) {
	var output []*mess.Message

	for _, message := range messages {
		// Skip control messages.
		if message.IsControl() {
			output = append(output, message)
			continue
		}

		var value float64
		result := message.Get(t.conf.Key)
		for i, res := range result.Array() {
			if i == 0 {
				value = res.Float()
				continue
			}

			switch t.conf.Operation {
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

		if err := message.Set(t.conf.SetKey, value); err != nil {
			return nil, fmt.Errorf("transform: proc_math: %v", err)
		}

		output = append(output, message)
	}

	return output, nil
}
