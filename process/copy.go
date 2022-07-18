package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/internal/json"
)

/*
Copy processes data by copying it. The processor supports these patterns:
	JSON:
	  	{"hello":"world"} >>> {"hello":"world","goodbye":"world"}
	from JSON:
  		{"hello":"world"} >>> world
	to JSON:
  		world >>> {"hello":"world"}

The processor uses this Jsonnet configuration:
	{
		type: 'copy',
		settings: {
			input_key: 'hello',
			output_key: 'goodbye',
		},
	}
*/
type Copy struct {
	Condition condition.OperatorConfig `json:"condition"`
	InputKey  string                   `json:"input_key"`
	OutputKey string                   `json:"output_key"`
}

// Slice processes a slice of bytes with the Copy processor. Conditions are optionally applied on the bytes to enable processing.
func (p Copy) Slice(ctx context.Context, s [][]byte) ([][]byte, error) {
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

// Byte processes bytes with the Copy processor.
func (p Copy) Byte(ctx context.Context, data []byte) ([]byte, error) {
	// JSON processing
	if p.InputKey != "" && p.OutputKey != "" {
		v := json.Get(data, p.InputKey)
		return json.Set(data, p.OutputKey, v)
	}

	// from JSON processing
	if p.InputKey != "" && p.OutputKey == "" {
		v := json.Get(data, p.InputKey)
		return []byte(v.String()), nil
	}

	// to JSON processing
	if p.InputKey == "" && p.OutputKey != "" {
		return json.Set([]byte{}, p.OutputKey, data)
	}

	return nil, fmt.Errorf("byter settings %+v: %w", p, ProcessorInvalidSettings)
}
