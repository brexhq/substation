package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/json"
)

// ConvertInvalidSettings is returned when the Convert processor is configured with invalid Input and Output settings.
const ConvertInvalidSettings = errors.Error("ConvertInvalidSettings")

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
	json:
		{"convert":"true"} >>> {"convert":true}
		{"convert":"-123"} >>> {"convert":-123}
		{"convert":123} >>> {"convert":"123"}
	json array:
		{"convert":["true","false"]} >>> {"convert":[true,false]}
		{"convert":["-123","-456"]} >>> {"convert":[-123,-456]}
		{"convert":[123,123.456]} >>> {"convert":["123","123.456"]}

The processor uses this Jsonnet configuration:
	{
		type: 'convert',
		settings: {
			input: {
				key: 'convert',
			},
			output: {
				key: 'convert',
			},
			options: {
				type: 'bool',
			}
		},
	}
*/
type Convert struct {
	Condition condition.OperatorConfig `json:"condition"`
	Input     string                   `json:"input"`
	Output    string                   `json:"output"`
	Options   ConvertOptions           `json:"options"`
}

// Slice processes a slice of bytes with the Convert processor. Conditions are optionally applied on the bytes to enable processing.
func (p Convert) Slice(ctx context.Context, s [][]byte) ([][]byte, error) {
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

// Byte processes bytes with the Convert processor.
func (p Convert) Byte(ctx context.Context, data []byte) ([]byte, error) {
	// json processing
	if p.Input != "" && p.Output != "" {
		value := json.Get(data, p.Input)
		if !value.IsArray() {
			c := p.convert(value)
			return json.Set(data, p.Output, c)
		}

		// json array processing
		var array []interface{}
		for _, v := range value.Array() {
			c := p.convert(v)
			array = append(array, c)
		}

		return json.Set(data, p.Output, array)
	}

	return nil, fmt.Errorf("byter settings %v: %v", p, ConvertInvalidSettings)
}

func (p Convert) convert(v json.Result) interface{} {
	switch t := p.Options.Type; t {
	case "bool":
		return v.Bool()
	case "int":
		return v.Int()
	case "float":
		return v.Float()
	case "uint":
		return v.Uint()
	case "string":
		return v.String()
	default:
		return nil
	}
}
