package process

import (
	"context"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/json"
)

// DeleteInvalidSettings is returned when the Copy processor is configured with invalid Input and Output settings.
const DeleteInvalidSettings = errors.Error("DeleteInvalidSettings")

/*
Delete processes data by deleting JSON keys. The processor supports these patterns:
	json:
  	{"hello":"world","goodbye":"world"} >>> {"hello":"world"}
*/
type Delete struct {
	Condition condition.OperatorConfig `mapstructure:"condition"`
	Input     Input                    `mapstructure:"input"`
}

// Channel processes a channel of byte slices with the Delete processor.
func (p Delete) Channel(ctx context.Context, ch <-chan []byte) (<-chan []byte, error) {
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

// Byte processes a byte slice with the Delete processor.
func (p Delete) Byte(ctx context.Context, object []byte) ([]byte, error) {
	// json processing
	if p.Input.Key != "" {
		return json.Delete(object, p.Input.Key)
	}

	return nil, DeleteInvalidSettings
}
