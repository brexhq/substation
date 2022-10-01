package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
)

/*
Drop processes data by "dropping" it -- the data is entirely removed and not emitted.

When loaded with a factory, the processor uses this JSON configuration:
	{
		type: "drop"
	}
*/
type Drop struct {
	Condition condition.Config `json:"condition"`
}

// ApplyBatch processes a slice of encapsulated data with the Drop processor. Conditions are optionally applied to the data to enable processing.
func (p Drop) ApplyBatch(ctx context.Context, caps []config.Capsule) ([]config.Capsule, error) {
	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, fmt.Errorf("drop applybatch: %v", err)
	}

	newCaps := newBatch(&caps)
	for _, cap := range caps {
		ok, err := op.Operate(ctx, cap)
		if err != nil {
			return nil, fmt.Errorf("drop applybatch: %v", err)
		}

		if !ok {
			newCaps = append(newCaps, cap)
			continue
		}
	}

	return newCaps, nil
}
