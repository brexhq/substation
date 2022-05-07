package process

import (
	"context"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/json"
)

// FlattenInvalidSettings is returned when the Flatten processor is configured with invalid Input and Output settings.
const FlattenInvalidSettings = errors.Error("FlattenInvalidSettings")

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
	json:
		{"flatten":["foo",["bar"]]} >>> {"flatten":["foo","bar"]}

The processor uses this Jsonnet configuration:
	{
		type: 'flatten',
		settings: {
			input: {
				key: 'flatten',
			},
			output: {
				key: 'flatten',
			},
		},
	}
*/
type Flatten struct {
	Condition condition.OperatorConfig `json:"condition"`
	Input     Input                    `json:"input"`
	Output    Output                   `json:"output"`
	Options   FlattenOptions           `json:"options"`
}

// Slice processes a slice of bytes with the Flatten processor. Conditions are optionally applied on the bytes to enable processing.
func (p Flatten) Slice(ctx context.Context, s [][]byte) ([][]byte, error) {
	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, err
	}

	slice := NewSlice(&s)
	for _, data := range s {
		ok, err := op.Operate(data)
		if err != nil {
			return nil, err
		}

		if !ok {
			slice = append(slice, data)
			continue
		}

		processed, err := p.Byte(ctx, data)
		if err != nil {
			return nil, err
		}
		slice = append(slice, processed)
	}

	return slice, nil
}

// Byte processes bytes with the Flatten processor.
func (p Flatten) Byte(ctx context.Context, data []byte) ([]byte, error) {
	// only supports json, so error early if there are no keys
	if p.Input.Key == "" && p.Output.Key == "" {
		return nil, FlattenInvalidSettings
	}

	var value json.Result
	if p.Options.Deep {
		value = json.Get(data, p.Input.Key+`|@flatten:{"deep":true}`)
	} else {
		value = json.Get(data, p.Input.Key+"|@flatten")
	}

	return json.Set(data, p.Output.Key, value)
}
