package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/config"
)

// flatten processes data by flattening object arrays.
//
// This processor supports the object handling pattern.
type _flatten struct {
	process
	Options _flattenOptions `json:"options"`
}

type _flattenOptions struct {
	// Deep determines if arrays should be deeply flattened.
	//
	// This is optional and defaults to false.
	Deep bool `json:"deep"`
}

// String returns the processor settings as an object.
func (p _flatten) String() string {
	return toString(p)
}

// Close closes resources opened by the processor.
func (p _flatten) Close(context.Context) error {
	return nil
}

// Batch processes one or more capsules with the processor. Conditions are
// optionally applied to the data to enable processing.
func (p _flatten) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	capsules, err := conditionalApply(ctx, capsules, p.Condition, p)
	if err != nil {
		return nil, fmt.Errorf("process _flatten: %v", err)
	}

	return capsules, nil
}

// Apply processes a capsule with the processor.
func (p _flatten) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	// only supports JSON, error early if there are no keys
	if p.Key == "" && p.SetKey == "" {
		return capsule, fmt.Errorf("process flatten: inputkey %s outputkey %s: %v", p.Key, p.SetKey, errInvalidDataPattern)
	}

	var value interface{}
	if p.Options.Deep {
		value = capsule.Get(p.Key + `|@flatten:{"deep":true}`)
	} else {
		value = capsule.Get(p.Key + `|@flatten`)
	}

	if err := capsule.Set(p.SetKey, value); err != nil {
		return capsule, fmt.Errorf("process flatten: %v", err)
	}

	return capsule, nil
}
