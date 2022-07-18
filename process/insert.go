package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/internal/json"
)

/*
InsertOptions contains custom options for the Insert processor:
	value:
		the value to insert
*/
type InsertOptions struct {
	Value interface{} `json:"value"`
}

/*
Insert processes data by inserting a value into a JSON object. The processor supports these patterns:
	JSON:
		{"foo":"bar"} >>> {"foo":"bar","baz":"qux"}

The processor uses this Jsonnet configuration:
	{
		type: 'insert',
		settings: {
			options: {
				value: 'qux',
			},
			output_key: 'baz',
		},
	}
*/
type Insert struct {
	Options   InsertOptions            `json:"options"`
	Condition condition.OperatorConfig `json:"condition"`
	OutputKey string                   `json:"output_key"`
}

// Slice processes a slice of bytes with the Insert processor. Conditions are optionally applied on the bytes to enable processing.
func (p Insert) Slice(ctx context.Context, s [][]byte) ([][]byte, error) {
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

// Byte processes bytes with the Insert processor.
func (p Insert) Byte(ctx context.Context, data []byte) ([]byte, error) {
	// only supports JSON, error early if there are no keys
	if p.OutputKey == "" {
		return nil, fmt.Errorf("byter settings %+v: %w", p, ProcessorInvalidSettings)
	}

	return json.Set(data, p.OutputKey, p.Options.Value)
}
