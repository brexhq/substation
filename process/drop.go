package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
)

type drop struct {
	process
}

// Close closes resources opened by the Drop processor.
func (p drop) Close(context.Context) error {
	return nil
}

func (p drop) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
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
