package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/config"
)

type insert struct {
	process
	Options insertOptions `json:"options"`
}

type insertOptions struct {
	Value interface{} `json:"value"`
}

// Close closes resources opened by the insert processor.
func (p insert) Close(context.Context) error {
	return nil
}

func (p insert) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	capsules, err := conditionalApply(ctx, capsules, p.Condition, p)
	if err != nil {
		return nil, fmt.Errorf("process insert: %v", err)
	}

	return capsules, nil
}

// Apply processes encapsulated data with the insert processor.
func (p insert) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	// only supports JSON, error early if there are no keys
	if p.SetKey == "" {
		return capsule, fmt.Errorf("process insert: outputkey %s: %v", p.SetKey, errInvalidDataPattern)
	}

	if err := capsule.Set(p.SetKey, p.Options.Value); err != nil {
		return capsule, fmt.Errorf("process insert: %v", err)
	}

	return capsule, nil
}
