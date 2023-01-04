package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/config"
)

// delete processes data by deleting keys from an object.
//
// This processor supports the object handling pattern.
type procDelete struct {
	process
}

// String returns the processor settings as an object.
func (p procDelete) String() string {
	return toString(p)
}

// Closes resources opened by the processor.
func (p procDelete) Close(context.Context) error {
	return nil
}

// Batch processes one or more capsules with the processor. Conditions are
// optionally applied to the data to enable processing.
func (p procDelete) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	return batchApply(ctx, capsules, p, p.Condition)
}

// Apply processes a capsule with the processor.
func (p procDelete) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	// only supports JSON, error early if there are no keys
	if p.Key == "" {
		return capsule, fmt.Errorf("process: delete: key %s: %v", p.Key, errInvalidDataPattern)
	}

	if err := capsule.Delete(p.Key); err != nil {
		return capsule, fmt.Errorf("process: delete: %v", err)
	}

	return capsule, nil
}
