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
type procDrop struct {
	process
}

// Create a new drop processor.
func newProcDrop(cfg config.Config) (p procDrop, err error) {
	if err = config.Decode(cfg.Settings, &p); err != nil {
		return procDrop{}, err
	}

	p.operator, err = condition.NewOperator(p.Condition)
	if err != nil {
		return procDrop{}, err
	}

	return p, nil
}

// String returns the processor settings as an object.
func (p procDrop) String() string {
	return toString(p)
}

// Closes resources opened by the processor.
func (p procDrop) Close(context.Context) error {
	return nil
}

// Batch processes one or more capsules with the processor. Conditions are
// optionally applied to the data to enable processing.
func (p procDrop) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	newCapsules := newBatch(&capsules)
	for _, capsule := range capsules {
		ok, err := p.operator.Operate(ctx, capsule)
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
