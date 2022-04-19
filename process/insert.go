package process

import (
	"context"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/internal/json"
)

/*
InsertOptions contain custom options settings for this processor.

Value: the value to insert.
*/
type InsertOptions struct {
	Value interface{} `mapstructure:"value"`
}

// Insert implements the Byter and Channeler interfaces and inserts a value into a JSON object. More information is available in the README.
type Insert struct {
	Condition condition.OperatorConfig `mapstructure:"condition"`
	Output    Output                   `mapstructure:"output"`
	Options   InsertOptions            `mapstructure:"options"`
}

// Channel processes a data channel of bytes with this processor. Conditions can be optionally applied on the channel data to enable processing.
func (p Insert) Channel(ctx context.Context, ch <-chan []byte) (<-chan []byte, error) {
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
func (p Insert) Byte(ctx context.Context, data []byte) ([]byte, error) {
	return json.Set(data, p.Output.Key, p.Options.Value)
}
