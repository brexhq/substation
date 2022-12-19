package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/config"
)

type copy struct {
	process
}

// Close closes resources opened by the Copy processor.
func (p copy) Close(context.Context) error {
	return nil
}

func (p copy) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	capsules, err := conditionalApply(ctx, capsules, p.Condition, p)
	if err != nil {
		return nil, fmt.Errorf("process capture: %v", err)
	}

	return capsules, nil
}

// Apply processes encapsulated data with the Copy processor.
func (p copy) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	// JSON processing
	if p.Key != "" && p.SetKey != "" {
		if err := capsule.Set(p.SetKey, capsule.Get(p.Key)); err != nil {
			return capsule, fmt.Errorf("process copy: %v", err)
		}

		return capsule, nil
	}

	// from JSON processing
	if p.Key != "" && p.SetKey == "" {
		result := capsule.Get(p.Key).String()

		capsule.SetData([]byte(result))
		return capsule, nil
	}

	// to JSON processing
	if p.Key == "" && p.SetKey != "" {
		if err := capsule.Set(p.SetKey, capsule.Data()); err != nil {
			return capsule, fmt.Errorf("process copy: %v", err)
		}

		return capsule, nil
	}

	return capsule, fmt.Errorf("process copy: inputkey %s outputkey %s: %w", p.Key, p.SetKey, errInvalidDataPattern)
}
