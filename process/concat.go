package process

import (
	"context"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/json"
)

// ConcatInvalidSettings is returned when the Concat processor is configured with invalid Input and Output settings.
const ConcatInvalidSettings = errors.Error("ConcatInvalidSettings")

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
	json:
		{"c1":"foo","c2":"bar"} >>> {"c3":"foo.bar"}
	json array:
		{"c1":["foo","baz"],"c2":["bar","qux"]} >>> {"c3":["foo.bar","baz.qux"]}

The processor uses this Jsonnet configuration:
	{
		type: 'concat',
		settings: {
			input: {
				keys: ['c1','c2'],
			},
			output: {
				key: 'c3',
			},
			options: {
				separator: '.',
			}
		},
	}
*/
type Concat struct {
	Condition condition.OperatorConfig `json:"condition"`
	Input     Inputs                   `json:"input"`
	Output    Output                   `json:"output"`
	Options   ConcatOptions            `json:"options"`
}

// Slice processes a slice of bytes with the Concat processor. Conditions are optionally applied on the bytes to enable processing.
func (p Concat) Slice(ctx context.Context, s [][]byte) ([][]byte, error) {
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

// Byte processes bytes with the Concat processor.
func (p Concat) Byte(ctx context.Context, data []byte) ([]byte, error) {
	if len(p.Input.Keys) != 0 && p.Output.Key != "" {
		count := len(p.Input.Keys) - 1

		cache := make(map[int]string)
		for i, key := range p.Input.Keys {
			value := json.Get(data, key)
			if value.Type.String() == "Null" {
				return data, nil
			}

			for x, v := range value.Array() {
				cache[x] += v.String()
				if i != count {
					cache[x] += p.Options.Separator
				}
			}
		}

		// json processing
		if len(cache) == 1 {
			return json.Set(data, p.Output.Key, cache[0])
		}

		// json array processing
		var array []string
		for _, v := range cache {
			array = append(array, v)
		}

		return json.Set(data, p.Output.Key, array)
	}

	return nil, ConcatInvalidSettings
}
