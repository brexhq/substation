package process

import (
	"context"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/json"
)

// ExpandInvalidSettings is returned when the Expand processor is configured with invalid Input and Output settings.
const ExpandInvalidSettings = errors.Error("ExpandInvalidSettings")

/*
ExpandOptions contains custom options settings for the Expand processor:
	Retain (optional):
		array of JSON keys to retain from the original object
*/
type ExpandOptions struct {
	Retain []string `mapstructure:"retain"` // retain fields found anywhere in input
}

/*
Expand processes data by creating individual events from objects in JSON arrays. The processor supports these patterns:
	json array:
		{"expand":[{"foo":"bar"}],"baz":"qux"} >>> {"foo":"bar","baz":"qux"}

The processor uses this Jsonnet configuration:
{
  type: 'expand',
  settings: {
    input: {
      key: 'expand',
    },
    options: {
      retain: ['baz'],
    }
  },
}
*/
type Expand struct {
	Condition condition.OperatorConfig `mapstructure:"condition"`
	Input     Input                    `mapstructure:"input"`
	Options   ExpandOptions            `mapstructure:"options"`
}

// Channel processes a data channel of byte slices with the Expand processor. Conditions are optionally applied on the channel data to enable processing.
func (p Expand) Channel(ctx context.Context, ch <-chan []byte) (<-chan []byte, error) {
	// only supports json, so error early if there is no input key
	if p.Input.Key == "" {
		return nil, ExpandInvalidSettings
	}

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

		// json array processing
		value := json.Get(data, p.Input.Key)
		for _, x := range value.Array() {
			var err error
			processed := []byte(x.String())
			for _, r := range p.Options.Retain {
				v := json.Get(data, r)
				processed, err = json.Set(processed, r, v)
				if err != nil {
					return nil, err
				}
			}

			array = append(array, processed)
		}
	}

	output := make(chan []byte, len(array))
	for _, x := range array {
		output <- x
	}
	defer close(output)
	return output, nil
}
