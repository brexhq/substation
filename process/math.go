package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/internal/json"
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
Math processes data by applying mathematic operations. The processor supports these patterns:
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

// Slice processes a slice of bytes with the Math processor. Conditions are optionally applied on the bytes to enable processing.
func (p Math) Slice(ctx context.Context, s [][]byte) ([][]byte, error) {
	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, fmt.Errorf("slicer settings %+v: %w", p, err)
	}

	slice := NewSlice(&s)
	for _, data := range s {
		ok, err := op.Operate(data)
		if err != nil {
			return nil, fmt.Errorf("slicer settings %+v: %w", p, err)
		}

		if !ok {
			slice = append(slice, data)
			continue
		}

		processed, err := p.Byte(ctx, data)
		if err != nil {
			return nil, fmt.Errorf("slicer: %v", err)
		}
		slice = append(slice, processed)
	}

	return slice, nil
}

// Byte processes bytes with the Math processor.
func (p Math) Byte(ctx context.Context, data []byte) ([]byte, error) {
	// error early if required options are missing
	if p.Options.Operation == "" {
		return nil, fmt.Errorf("byter settings %+v: %w", p, ProcessorInvalidSettings)
	}

	// only supports JSON, error early if there are no keys
	if p.InputKey == "" || p.OutputKey == "" {
		return nil, fmt.Errorf("byter settings %+v: %w", p, ProcessorInvalidSettings)
	}

	var tmp int64
	value := json.Get(data, p.InputKey)
	for x, v := range value.Array() {
		if x == 0 {
			tmp = v.Int()
			continue
		}

		switch p.Options.Operation {
		case "add":
			tmp += v.Int()
		case "subtract":
			tmp -= v.Int()
		case "multiply":
			tmp = tmp * v.Int()
		case "divide":
			tmp = tmp / v.Int()
		}
	}

	return json.Set(data, p.OutputKey, tmp)
}
