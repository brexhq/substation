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
	Deep bool `mapstructure:"deep"`
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
	Condition condition.OperatorConfig `mapstructure:"condition"`
	Input     Input                    `mapstructure:"input"`
	Output    Output                   `mapstructure:"output"`
	Options   FlattenOptions           `mapstructure:"options"`
}

// Channel processes a data channel of byte slices with the Flatten processor. Conditions are optionally applied on the channel data to enable processing.
func (p Flatten) Channel(ctx context.Context, ch <-chan []byte) (<-chan []byte, error) {
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

// Byte processes a byte slice with the Flatten processor.
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
