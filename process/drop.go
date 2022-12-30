package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
)

// drop processes data by removing and not emitting it.
//
// This processor supports the data and object handling patterns.
type _drop struct {
	process
}

// String returns the processor settings as an object.
func (p _drop) String() string {
	return toString(p)
}

// Close closes resources opened by the processor.
func (p _drop) Close(context.Context) error {
	return nil
}

// Batch processes one or more capsules with the processor. Conditions are
// optionally applied to the data to enable processing.
func (p _drop) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	op, err := condition.MakeOperator(p.Condition)
	if err != nil {
		return nil, fmt.Errorf("process: drop: %v", err)
	}

	newCapsules := newBatch(&capsules)
	for _, capsule := range capsules {
		ok, err := op.Operate(ctx, capsule)
		if err != nil {
			return nil, fmt.Errorf("process: drop: %v", err)
		}

		if !ok {
			newCapsules = append(newCapsules, capsule)
			continue
		}
	}

	return newCapsules, nil
}
