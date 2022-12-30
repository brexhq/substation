package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/config"
)

// insert processes data by inserting a value into an object.
//
// This processor supports the object handling pattern.
type _insert struct {
	process
	Options _insertOptions `json:"options"`
}

type _insertOptions struct {
	// Value inserted into the object.
	Value interface{} `json:"value"`
}

// String returns the processor settings as an object.
func (p _insert) String() string {
	return toString(p)
}

// Close closes resources opened by the processor.
func (p _insert) Close(context.Context) error {
	return nil
}

// Batch processes one or more capsules with the processor. Conditions are
// optionally applied to the data to enable processing.
func (p _insert) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	return batchApply(ctx, capsules, p, p.Condition)
}

// Apply processes a capsule with the processor.
func (p _insert) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	// only supports JSON, error early if there are no keys
	if p.SetKey == "" {
		return capsule, fmt.Errorf("process: insert: set_key %s: %v", p.SetKey, errInvalidDataPattern)
	}

	if err := capsule.Set(p.SetKey, p.Options.Value); err != nil {
		return capsule, fmt.Errorf("process: insert: %v", err)
	}

	return capsule, nil
}
