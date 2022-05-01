package process

import (
	"context"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/json"
)

// CopyInvalidSettings is returned when the Copy processor is configured with invalid Input and Output settings.
const CopyInvalidSettings = errors.Error("CopyInvalidSettings")

/*
Copy processes data by copying it. The processor supports these patterns:
	json:
  	{"hello":"world"} >>> {"hello":"world","goodbye":"world"}
	from json:
  	{"hello":"world"} >>> world
	to json:
  	world >>> {"hello":"world"}
*/
type Copy struct {
	Condition condition.OperatorConfig `mapstructure:"condition"`
	Input     Input                    `mapstructure:"input"`
	Output    Output                   `mapstructure:"output"`
}

// Channel processes a channel of byte slices with the Copy processor. Conditions are optionally applied on the channel data to enable processing.
func (p Copy) Channel(ctx context.Context, ch <-chan []byte) (<-chan []byte, error) {
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

// Byte processes a byte slice with the Copy processor.
func (p Copy) Byte(ctx context.Context, data []byte) ([]byte, error) {
	// JSON processing
	if p.Input.Key != "" && p.Output.Key != "" {
		v := json.Get(data, p.Input.Key)
		return json.Set(data, p.Output.Key, v)
	}

	// from JSON processing
	if p.Input.Key != "" && p.Output.Key == "" {
		v := json.Get(data, p.Input.Key)
		return []byte(v.String()), nil
	}

	// to JSON processing
	if p.Input.Key == "" && p.Output.Key != "" {
		return json.Set([]byte(""), p.Output.Key, data)
	}

	return nil, CopyInvalidSettings
}
