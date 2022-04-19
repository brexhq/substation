package process

import (
	"context"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/internal/json"
)

/*
ExpandOptions contain custom options settings for this processor.

Retain: array of JSON keys to retain from the original object and insert into the new objects
*/
type ExpandOptions struct {
	Retain []string `mapstructure:"retain"` // retain fields found anywhere in input
}

// Expand implements the Channeler interface and expands data in JSON arrays into individual events. More information is available in the README.
type Expand struct {
	Condition condition.OperatorConfig `mapstructure:"condition"`
	Input     Input                    `mapstructure:"input"`
	Options   ExpandOptions            `mapstructure:"options"`
}

// Channel processes a data channel of bytes with this processor. Conditions can be optionally applied on the channel data to enable processing.
func (p Expand) Channel(ctx context.Context, ch <-chan []byte) (<-chan []byte, error) {
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

		if len(p.Input.Key) == 0 {
			p.Input.Key = "@this"
		}

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
