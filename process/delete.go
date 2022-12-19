package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/config"
)

type delete struct {
	process
}

// Close closes resources opened by the Delete processor.
func (p delete) Close(context.Context) error {
	return nil
}

func (p delete) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	capsules, err := conditionalApply(ctx, capsules, p.Condition, p)
	if err != nil {
		return nil, fmt.Errorf("process capture: %v", err)
	}

	return capsules, nil
}

// Apply processes encapsulated data with the Delete processor.
func (p delete) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	// only supports JSON, error early if there are no keys
	if p.Key == "" {
		return capsule, fmt.Errorf("process delete: inputkey %s: %v", p.Key, errInvalidDataPattern)
	}

	if err := capsule.Delete(p.Key); err != nil {
		return capsule, fmt.Errorf("process delete: %v", err)
	}

	return capsule, nil
}
