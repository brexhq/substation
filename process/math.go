package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
)

/*
MathOptions contains custom options for the Math processor:
	Operation:
		the operator applied to the data
		must be one of:
			add
			subtract
			multiply
			divide
*/
type MathOptions struct {
	Operation string `json:"operation"`
}

/*
Math processes encapsulated data by applying mathematic operations. The processor supports these patterns:
	JSON:
		{"math":[1,3]} >>> {"math":4}

The processor uses this Jsonnet configuration:
	{
		type: 'math',
		settings: {
			options: {
				operation: 'add',
			},
			input_key: 'math',
			output_key: 'math',
		},
	}
*/
type Math struct {
	Options   MathOptions              `json:"options"`
	Condition condition.OperatorConfig `json:"condition"`
	InputKey  string                   `json:"input_key"`
	OutputKey string                   `json:"output_key"`
}

// ApplyBatch processes a slice of encapsulated data with the Math processor. Conditions are optionally applied to the data to enable processing.
func (p Math) ApplyBatch(ctx context.Context, caps []config.Capsule) ([]config.Capsule, error) {
	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, fmt.Errorf("applybatch settings %+v: %w", p, err)
	}

	caps, err = conditionallyApplyBatch(ctx, caps, op, p)
	if err != nil {
		return nil, fmt.Errorf("applybatch settings %+v: %w", p, err)
	}

	return caps, nil
}

// Apply processes encapsulated data with the Math processor.
func (p Math) Apply(ctx context.Context, cap config.Capsule) (config.Capsule, error) {
	// error early if required options are missing
	if p.Options.Operation == "" {
		return cap, fmt.Errorf("applicator settings %+v: %w", p, ProcessorInvalidSettings)
	}

	// only supports JSON, error early if there are no keys
	if p.InputKey == "" && p.OutputKey == "" {
		return cap, fmt.Errorf("applicator settings %+v: %w", p, ProcessorInvalidSettings)
	}

	var tmp int64
	res := cap.Get(p.InputKey)
	for idx, val := range res.Array() {
		if idx == 0 {
			tmp = val.Int()
			continue
		}

		switch p.Options.Operation {
		case "add":
			tmp += val.Int()
		case "subtract":
			tmp -= val.Int()
		case "multiply":
			tmp = tmp * val.Int()
		case "divide":
			tmp = tmp / val.Int()
		}
	}

	cap.Set(p.OutputKey, tmp)
	return cap, nil
}
