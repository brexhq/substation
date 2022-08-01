package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/internal/json"
)

/*
FlattenOptions contains custom options settings for the Flatten processor:
	Deep (optional):
		deeply flattens nested arrays
*/
type FlattenOptions struct {
	Deep bool `json:"deep"`
}

/*
Flatten processes data by flattening JSON arrays. The processor supports these patterns:
	JSON:
		{"flatten":["foo",["bar"]]} >>> {"flatten":["foo","bar"]}

The processor uses this Jsonnet configuration:
	{
		type: 'flatten',
		settings: {
			input_key: 'flatten',
			output_key: 'flatten',
		},
	}
*/
type Flatten struct {
	Options   FlattenOptions           `json:"options"`
	Condition condition.OperatorConfig `json:"condition"`
	InputKey  string                   `json:"input_key"`
	OutputKey string                   `json:"output_key"`
}

// Slice processes a slice of bytes with the Flatten processor. Conditions are optionally applied on the bytes to enable processing.
func (p Flatten) Slice(ctx context.Context, s [][]byte) ([][]byte, error) {
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

// Byte processes bytes with the Flatten processor.
func (p Flatten) Byte(ctx context.Context, data []byte) ([]byte, error) {
	// only supports JSON, error early if there are no keys
	if p.InputKey == "" || p.OutputKey == "" {
		return nil, fmt.Errorf("byter settings %+v: %w", p, ProcessorInvalidSettings)
	}

	var value json.Result
	if p.Options.Deep {
		value = json.Get(data, p.InputKey+`|@flatten:{"deep":true}`)
	} else {
		value = json.Get(data, p.InputKey+"|@flatten")
	}

	return json.Set(data, p.OutputKey, value)
}
