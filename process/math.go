package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/config"
)

// math processes data by applying mathematic operations.
//
// This processor supports the object handling pattern.
type _math struct {
	process
	Options _mathOptions `json:"options"`
}

type _mathOptions struct {
	// Operation determines the operator applied to the data.
	//
	// Must be one of:
	//
	// - add
	//
	// - subtract
	//
	// - multiply
	//
	// - divide
	Operation string `json:"operation"`
}

// String returns the processor settings as an object.
func (p _math) String() string {
	return toString(p)
}

// Close closes resources opened by the processor.
func (p _math) Close(context.Context) error {
	return nil
}

// Batch processes one or more capsules with the processor. Conditions are
// optionally applied to the data to enable processing.
func (p _math) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	capsules, err := conditionalApply(ctx, capsules, p.Condition, p)
	if err != nil {
		return nil, fmt.Errorf("process math: %v", err)
	}

	return capsules, nil
}

// Apply processes a capsule with the processor.
func (p _math) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	// error early if required options are missing
	if p.Options.Operation == "" {
		return capsule, fmt.Errorf("process math: options %+v: %v", p.Options, errMissingRequiredOptions)
	}

	// only supports JSON, error early if there are no keys
	if p.Key == "" && p.SetKey == "" {
		return capsule, fmt.Errorf("process math: inputkey %s outputkey %s: %v", p.Key, p.SetKey, errInvalidDataPattern)
	}

	var value int64
	result := capsule.Get(p.Key)
	for i, res := range result.Array() {
		if i == 0 {
			value = res.Int()
			continue
		}

		switch p.Options.Operation {
		case "add":
			value += res.Int()
		case "subtract":
			value -= res.Int()
		case "multiply":
			value *= res.Int()
		case "divide":
			value /= res.Int()
		}
	}

	if err := capsule.Set(p.SetKey, value); err != nil {
		return capsule, fmt.Errorf("process math: %v", err)
	}

	return capsule, nil
}
