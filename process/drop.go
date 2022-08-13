package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
)

/*
Drop processes encapsulated data by dropping it from a data channel. The processor uses this Jsonnet configuration:
	{
		type: 'drop',
	}
*/
type Drop struct {
	Condition condition.OperatorConfig `json:"condition"`
}

// ApplyBatch processes a slice of encapsulated data with the Drop processor. Conditions are optionally applied to the data to enable processing.
func (p Drop) ApplyBatch(ctx context.Context, caps []config.Capsule) ([]config.Capsule, error) {
	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, fmt.Errorf("applybatch settings %+v: %w", p, err)
	}

	newCaps := NewBatch(&caps)
	for _, cap := range caps {
		ok, err := op.Operate(cap)
		if err != nil {
			return nil, fmt.Errorf("applybatch settings %+v: %w", p, err)
		}

		if !ok {
			newCaps = append(newCaps, cap)
			continue
		}
	}

	return newCaps, nil
}
