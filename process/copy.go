package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/config"
)

// copy processes data by copying it into, from, and inside objects.
//
// This processor supports the data and object handling patterns.
type _copy struct {
	process
}

// String returns the processor settings as an object.
func (p _copy) String() string {
	return toString(p)
}

// Close closes resources opened by the processor.
func (p _copy) Close(context.Context) error {
	return nil
}

// Batch processes one or more capsules with the processor. Conditions are
// optionally applied to the data to enable processing.
func (p _copy) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	return conditionalApply(ctx, capsules, p.Condition, p)
}

// Apply processes a capsule with the processor.
func (p _copy) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
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
