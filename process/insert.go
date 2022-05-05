package process

import (
	"context"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/json"
)

// InsertInvalidSettings is returned when the Insert processor is configured with invalid Input and Output settings.
const InsertInvalidSettings = errors.Error("InsertInvalidSettings")

/*
InsertOptions contains custom options for the Insert processor:
	value:
		the value to insert
*/
type InsertOptions struct {
	Value interface{} `mapstructure:"value"`
}

/*
Insert processes data by inserting a value into a JSON object. The processor supports these patterns:
	json:
		{"foo":"bar"} >>> {"foo":"bar","baz":"qux"}

The processor uses this Jsonnet configuration:
	{
		type: 'insert',
		settings: {
			output: {
				key: 'baz',
			}
			options: {
				value: 'qux',
			}
		},
	}
*/
type Insert struct {
	Condition condition.OperatorConfig `mapstructure:"condition"`
	Output    Output                   `mapstructure:"output"`
	Options   InsertOptions            `mapstructure:"options"`
}

// Channel processes a data channel of byte slices with the Insert processor. Conditions are optionally applied on the channel data to enable processing.
func (p Insert) Channel(ctx context.Context, ch <-chan []byte) (<-chan []byte, error) {
	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, err
	}

	var array [][]byte
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

// Byte processes a byte slice with the Insert processor.
func (p Insert) Byte(ctx context.Context, data []byte) ([]byte, error) {
	// json processing
	if p.Output.Key != "" {
		return json.Set(data, p.Output.Key, p.Options.Value)
	}

	return nil, InsertInvalidSettings
}
