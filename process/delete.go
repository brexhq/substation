package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/config"
)

// delete processes data by deleting keys from an object.
//
// This processor supports the object handling pattern.
type _delete struct {
	process
}

// String returns the processor settings as an object.
func (p _delete) String() string {
	return toString(p)
}

// Close closes resources opened by the processor.
func (p _delete) Close(context.Context) error {
	return nil
}

// Batch processes one or more capsules with the processor. Conditions are
// optionally applied to the data to enable processing.
func (p _delete) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	return batchApply(ctx, capsules, p, p.Condition)
}

// Apply processes a capsule with the processor.
func (p _delete) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	// only supports JSON, error early if there are no keys
	if p.Key == "" {
		return capsule, fmt.Errorf("process delete: inputkey %s: %v", p.Key, errInvalidDataPattern)
	}

	if err := capsule.Delete(p.Key); err != nil {
		return capsule, fmt.Errorf("process delete: %v", err)
	}

	return capsule, nil
}
