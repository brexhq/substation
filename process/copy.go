package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/json"
)

// CopyInvalidSettings is returned when the Copy processor is configured with invalid Input and Output settings.
const CopyInvalidSettings = errors.Error("CopyInvalidSettings")

/*
Copy processes data by copying it. The processor supports these patterns:
	json:
	  	{"hello":"world"} >>> {"hello":"world","goodbye":"world"}
	from json:
  		{"hello":"world"} >>> world
	to json:
  		world >>> {"hello":"world"}

The processor uses this Jsonnet configuration:
	{
		type: 'copy',
		settings: {
			input: {
				key: 'hello',
			},
			output: {
				key: 'goodbye',
},
	}
*/
type Copy struct {
	Condition condition.OperatorConfig `json:"condition"`
	Input     string                   `json:"input"`
	Output    string                   `json:"output"`
}

// Slice processes a slice of bytes with the Copy processor. Conditions are optionally applied on the bytes to enable processing.
func (p Copy) Slice(ctx context.Context, s [][]byte) ([][]byte, error) {
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

// Byte processes bytes with the Copy processor.
func (p Copy) Byte(ctx context.Context, data []byte) ([]byte, error) {
	// json processing
	if p.Input != "" && p.Output != "" {
		v := json.Get(data, p.Input)
		return json.Set(data, p.Output, v)
	}

	// from json processing
	if p.Input != "" && p.Output == "" {
		v := json.Get(data, p.Input)
		return []byte(v.String()), nil
	}

	// to json processing
	if p.Input == "" && p.Output != "" {
		return json.Set([]byte(""), p.Output, data)
	}

	return nil, fmt.Errorf("byter settings %v: %v", p, CopyInvalidSettings)
}
