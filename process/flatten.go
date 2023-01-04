package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/config"
)

// flatten processes data by flattening object arrays.
//
// This processor supports the object handling pattern.
type procFlatten struct {
	process
	Options procFlattenOptions `json:"options"`
}

type procFlattenOptions struct {
	// Deep determines if arrays should be deeply flattened.
	//
	// This is optional and defaults to false.
	Deep bool `json:"deep"`
}

// String returns the processor settings as an object.
func (p procFlatten) String() string {
	return toString(p)
}

// Closes resources opened by the processor.
func (p procFlatten) Close(context.Context) error {
	return nil
}

// Batch processes one or more capsules with the processor. Conditions are
// optionally applied to the data to enable processing.
func (p procFlatten) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	return batchApply(ctx, capsules, p, p.Condition)
}

// Apply processes a capsule with the processor.
func (p procFlatten) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	// only supports JSON, error early if there are no keys
	if p.Key == "" && p.SetKey == "" {
		return capsule, fmt.Errorf("process: flatten: key %s set_key %s: %v", p.Key, p.SetKey, errInvalidDataPattern)
	}

	var value interface{}
	if p.Options.Deep {
		value = capsule.Get(p.Key + `|@flatten:{"deep":true}`)
	} else {
		value = capsule.Get(p.Key + `|@flatten`)
	}

	if err := capsule.Set(p.SetKey, value); err != nil {
		return capsule, fmt.Errorf("process: flatten: %v", err)
	}

	return capsule, nil
}
