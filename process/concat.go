package process

import (
	"context"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/internal/json"
)

/*
ConcatOptions contain custom options settings for this processor.

Separator: the string that separates the concatenated values.
*/
type ConcatOptions struct {
	Separator string `mapstructure:"separator"`
}

// Concat implements the Byter and Channeler interfaces and concatenates multiple JSON keys into a single value with a separator character. More information is available in the README.
type Concat struct {
	Condition condition.OperatorConfig `mapstructure:"condition"`
	Input     Inputs                   `mapstructure:"input"`
	Output    Output                   `mapstructure:"output"`
	Options   ConcatOptions            `mapstructure:"options"`
}

// Channel processes a data channel of bytes with this processor. Conditions can be optionally applied on the channel data to enable processing.
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

// Byte processes a byte slice with this processor
func (p Concat) Byte(ctx context.Context, data []byte) ([]byte, error) {
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

	var array []string
	for _, v := range cache {
		array = append(array, v)
	}

	if len(array) == 1 {
		return json.Set(data, p.Output.Key, array[0])
	}
	return json.Set(data, p.Output.Key, array)
}
