package process

import (
	"context"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/internal/json"
)

/*
MathOptions contain custom options settings for this processor.

Operation: the math operation applied to the data.
*/
type MathOptions struct {
	Operation string `mapstructure:"operation"`
}

// Math implements the Byter and Channeler interfaces and applies mathematical operations to data. More information is available in the README.
type Math struct {
	Condition condition.OperatorConfig `mapstructure:"condition"`
	Input     Inputs                   `mapstructure:"input"`
	Output    Output                   `mapstructure:"output"`
	Options   MathOptions              `mapstructure:"options"`
}

// Channel processes a data channel of bytes with this processor. Conditions can be optionally applied on the channel data to enable processing.
func (p Math) Channel(ctx context.Context, ch <-chan []byte) (<-chan []byte, error) {
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
func (p Math) Byte(ctx context.Context, data []byte) ([]byte, error) {
	cache := make(map[int]int64)
	for i, key := range p.Input.Keys {
		value := json.Get(data, key)

		for x, v := range value.Array() {
			if i == 0 {
				cache[x] = v.Int()
				continue
			}

			switch p.Options.Operation {
			case "add":
				cache[x] = cache[x] + v.Int()
			case "subtract":
				cache[x] = cache[x] - v.Int()
			}
		}
	}

	var array []int64
	for _, v := range cache {
		array = append(array, v)
	}

	if len(array) == 1 {
		return json.Set(data, p.Output.Key, array[0])
	}
	return json.Set(data, p.Output.Key, array)
}
