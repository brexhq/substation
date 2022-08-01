package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/internal/json"
)

/*
ConvertOptions contains custom options for the Convert processor:
	Type:
		the type that the value should be converted to
		must be one of:
			bool (boolean)
			int (integer)
			float
			uint (unsigned integer)
			string
*/
type ConvertOptions struct {
	Type string `json:"type"`
}

/*
Convert processes data by converting values between types (e.g., string to integer, integer to float). The processor supports these patterns:
	JSON:
		{"convert":"true"} >>> {"convert":true}
		{"convert":"-123"} >>> {"convert":-123}
		{"convert":123} >>> {"convert":"123"}

The processor uses this Jsonnet configuration:
	{
		type: 'convert',
		settings: {
			options: {
				type: 'bool',
			},
			input_key: 'convert',
			output_key: 'convert',
		},
	}
*/
type Convert struct {
	Options   ConvertOptions           `json:"options"`
	Condition condition.OperatorConfig `json:"condition"`
	InputKey  string                   `json:"input_key"`
	OutputKey string                   `json:"output_key"`
}

// Slice processes a slice of bytes with the Convert processor. Conditions are optionally applied on the bytes to enable processing.
func (p Convert) Slice(ctx context.Context, s [][]byte) ([][]byte, error) {
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

// Byte processes bytes with the Convert processor.
func (p Convert) Byte(ctx context.Context, data []byte) ([]byte, error) {
	// error early if required options are missing
	if p.Options.Type == "" {
		return nil, fmt.Errorf("byter settings %+v: %w", p, ProcessorInvalidSettings)
	}

	// only supports JSON, error early if there are no keys
	if p.InputKey == "" || p.OutputKey == "" {
		return nil, fmt.Errorf("byter settings %+v: %w", p, ProcessorInvalidSettings)
	}

	value := json.Get(data, p.InputKey)
	switch p.Options.Type {
	case "bool":
		return json.Set(data, p.OutputKey, value.Bool())
	case "int":
		return json.Set(data, p.OutputKey, value.Int())
	case "float":
		return json.Set(data, p.OutputKey, value.Float())
	case "uint":
		return json.Set(data, p.OutputKey, value.Uint())
	case "string":
		return json.Set(data, p.OutputKey, value.String())
	}

	return nil, fmt.Errorf("byter settings %+v: %w", p, ProcessorInvalidSettings)
}
