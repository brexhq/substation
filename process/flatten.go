package process

import (
	"context"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/internal/json"
)

/*
FlattenOptions contain custom options settings for this processor.

Deep: deeply flattens nested arrays.
*/
type FlattenOptions struct {
	Deep bool `mapstructure:"deep"`
}

// Flatten implements the Byter and Channeler interfaces and flattens JSON arrays. More information is available in the README.
type Flatten struct {
	Condition condition.OperatorConfig `mapstructure:"condition"`
	Input     Input                    `mapstructure:"input"`
	Output    Output                   `mapstructure:"output"`
	Options   FlattenOptions           `mapstructure:"options"`
}

// Channel processes a data channel of bytes with this processor. Conditions can be optionally applied on the channel data to enable processing.
func (p Flatten) Channel(ctx context.Context, ch <-chan []byte) (<-chan []byte, error) {
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
func (p Flatten) Byte(ctx context.Context, data []byte) ([]byte, error) {
	var value json.Result
	if p.Options.Deep {
		value = json.Get(data, p.Input.Key+"|@flatten:deep")
	} else {
		value = json.Get(data, p.Input.Key+"|@flatten")
	}

	return json.Set(data, p.Output.Key, value)
}
