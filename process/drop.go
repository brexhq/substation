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
	Condition condition.OperatorConfig `json:"condition"`
}

// Slice processes a slice of bytes with the Drop processor. Conditions are optionally applied on the bytes to enable processing.
func (p Drop) Slice(ctx context.Context, s [][]byte) ([][]byte, error) {
	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, err
	}

	slice := NewSlice(&s)
	for _, data := range s {
		ok, err := op.Operate(data)
		if err != nil {
			return nil, err
		}

		if !ok {
			slice = append(slice, data)
			continue
		}
	}

	return slice, nil
}
