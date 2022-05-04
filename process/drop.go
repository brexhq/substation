package process

import (
	"context"

	"github.com/brexhq/substation/condition"
)

/*
Drop processes data by dropping it from a data channel. The processor uses this Jsonnet configuration:
	{
		type: 'drop',
	}
*/
type Drop struct {
	Condition condition.OperatorConfig `mapstructure:"condition"`
}

// Channel processes a data channel of byte slices with the Drop processor. Conditions are optionally applied on the channel data to enable processing.
func (p Drop) Channel(ctx context.Context, ch <-chan []byte) (<-chan []byte, error) {
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
	}

	output := make(chan []byte, len(array))
	for _, x := range array {
		output <- x
	}
	close(output)
	return output, nil
}
