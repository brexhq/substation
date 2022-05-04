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
	separator:
		the string that separates the concatenated values
*/
type ConcatOptions struct {
	Separator string `mapstructure:"separator"`
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
	Condition condition.OperatorConfig `mapstructure:"condition"`
	Input     Inputs                   `mapstructure:"input"`
	Output    Output                   `mapstructure:"output"`
	Options   ConcatOptions            `mapstructure:"options"`
}

// Channel processes a data channel of byte slices with the Concat processor. Conditions are optionally applied on the channel data to enable processing.
func (p Concat) Channel(ctx context.Context, ch <-chan []byte) (<-chan []byte, error) {
	var array [][]byte

	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, err
	}

	for data := range ch {
		ok, err := op.Operate(data)
		if err != nil {
			return nil, err
		}

		if !ok {
			array = append(array, data)
			continue
		}

		processed, err := p.Byte(ctx, data)
		if err != nil {
			return nil, err
		}
		array = append(array, processed)
	}

	output := make(chan []byte, len(array))
	for _, x := range array {
		output <- x
	}
	close(output)
	return output, nil

}

// Byte processes a byte slice with the Concat processor.
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
