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

// Close closes resources opened by the Drop processor.
func (p Drop) Close(context.Context) error {
	return nil
}

// ApplyBatch processes a slice of encapsulated data with the Drop processor. Conditions are optionally applied to the data to enable processing.
func (p Drop) ApplyBatch(ctx context.Context, capsules []config.Capsule) ([]config.Capsule, error) {
	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, fmt.Errorf("process drop: %v", err)
	}

	newCapsules := newBatch(&capsules)
	for _, capsule := range capsules {
		ok, err := op.Operate(ctx, capsule)
		if err != nil {
			return nil, fmt.Errorf("process drop: %v", err)
		}

		if !ok {
			newCapsules = append(newCapsules, capsule)
			continue
		}
	}

	return newCapsules, nil
}
