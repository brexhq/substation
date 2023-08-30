package transform

import (
	"context"
	gojson "encoding/json"
	"fmt"

	"golang.org/x/exp/slices"

	"github.com/brexhq/substation/config"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/message"
)

type modMathConfig struct {
	Object configObject `json:"object"`
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

type modMath struct {
	conf modMathConfig
}

func newModMath(_ context.Context, cfg config.Config) (*modMath, error) {
	conf := modMathConfig{}
	if err := iconfig.Decode(cfg.Settings, &conf); err != nil {
		return nil, fmt.Errorf("transform: new_mod_math: %v", err)
	}

	// Validate required options.
	if conf.Object.Key == "" && conf.Object.SetKey != "" {
		return nil, fmt.Errorf("transform: new_mod_math: object_key: %v", errors.ErrMissingRequiredOption)
	}

	if conf.Object.Key != "" && conf.Object.SetKey == "" {
		return nil, fmt.Errorf("transform: new_mod_math: object_set_key: %v", errors.ErrMissingRequiredOption)
	}

	if conf.Operation == "" {
		return nil, fmt.Errorf("transform: new_mod_math: operation: %v", errors.ErrMissingRequiredOption)
	}

	if !slices.Contains(
		[]string{
			"add",
			"subtract",
			"multiply",
			"divide",
		},
		conf.Operation) {
		return nil, fmt.Errorf("transform: new_mod_math: operation %q: %v", conf.Operation, errors.ErrInvalidOption)
	}

	tf := modMath{
		conf: conf,
	}

	return &tf, nil
}

func (tf *modMath) String() string {
	b, _ := gojson.Marshal(tf.conf)
	return string(b)
}

func (*modMath) Close(context.Context) error {
	return nil
}

func (tf *modMath) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	// Skip interrupt messages.
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	result := msg.GetObject(tf.conf.Object.Key).Array()
	if len(result) <= 1 {
		return []*message.Message{msg}, nil
	}

	var value float64
	for i, res := range result {
		if i == 0 {
			value = res.Float()
			continue
		}

		switch tf.conf.Operation {
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

	if err := msg.SetObject(tf.conf.Object.SetKey, value); err != nil {
		return nil, fmt.Errorf("transform: mod_math: %v", err)
	}

	return []*message.Message{msg}, nil
}
