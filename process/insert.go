package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/json"
)

// InsertInvalidSettings is returned when the Insert processor is configured with invalid Input and Output settings.
const InsertInvalidSettings = errors.Error("InsertInvalidSettings")

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
	json:
		{"foo":"bar"} >>> {"foo":"bar","baz":"qux"}

The processor uses this Jsonnet configuration:
	{
		type: 'insert',
		settings: {
			output: {
				key: 'baz',
			}
			options: {
				value: 'qux',
			}
		},
	}
*/
type Insert struct {
	Condition condition.OperatorConfig `json:"condition"`
	Output    Output                   `json:"output"`
	Options   InsertOptions            `json:"options"`
}

// Slice processes a slice of bytes with the Insert processor. Conditions are optionally applied on the bytes to enable processing.
func (p Insert) Slice(ctx context.Context, s [][]byte) ([][]byte, error) {
	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, fmt.Errorf("slicer settings %v: %v", p, err)
	}

	slice := NewSlice(&s)
	for _, data := range s {
		ok, err := op.Operate(data)
		if err != nil {
			return nil, fmt.Errorf("slicer settings %v: %v", p, err)
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
	// json processing
	if p.Output.Key != "" {
		return json.Set(data, p.Output.Key, p.Options.Value)
	}

	return nil, fmt.Errorf("byter settings %v: %v", p, InsertInvalidSettings)
}
