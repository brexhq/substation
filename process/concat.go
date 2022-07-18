package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/internal/json"
)

/*
ConcatOptions contains custom options for the Concat processor:
	Separator:
		the string that separates the concatenated values
*/
type ConcatOptions struct {
	Separator string `json:"separator"`
}

/*
Concat processes data by concatenating multiple values together with a separator. The processor supports these patterns:
	JSON:
		{"concat":["foo","bar"]} >>> {"concat":"foo.bar"}

The processor uses this Jsonnet configuration:
	{
		type: 'concat',
		settings: {
			input_key: 'concat',
			output_key: 'concat',
			options: {
				separator: '.',
			}
		},
	}
*/
type Concat struct {
	Options   ConcatOptions            `json:"options"`
	Condition condition.OperatorConfig `json:"condition"`
	InputKey  string                   `json:"input_key"`
	OutputKey string                   `json:"output_key"`
}

// Slice processes a slice of bytes with the Concat processor. Conditions are optionally applied on the bytes to enable processing.
func (p Concat) Slice(ctx context.Context, s [][]byte) ([][]byte, error) {
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

// Byte processes bytes with the Concat processor.
func (p Concat) Byte(ctx context.Context, data []byte) ([]byte, error) {
	// error early if required options are missing
	if p.Options.Separator == "" {
		return nil, fmt.Errorf("byter settings %+v: %v", p, ProcessorInvalidSettings)
	}

	// only supports JSON, error early if there are no keys
	if p.InputKey == "" && p.OutputKey == "" {
		return nil, fmt.Errorf("byter settings %v: %v", p, ProcessorInvalidSettings)
	}

	// data is processed by retrieving and iterating the
	// array (InputKey) containing string values and joining
	// each one with the separator string
	//
	// root:
	// 	{"concat":["foo","bar","baz"]}
	// concatenated:
	// 	{"concat:"foo.bar.baz"}
	var tmp string
	value := json.Get(data, p.InputKey)
	for idx, val := range value.Array() {
		tmp += val.String()
		if idx != len(value.Array())-1 {
			tmp += p.Options.Separator
		}
	}

	return json.Set(data, p.OutputKey, tmp)
}
