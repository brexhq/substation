package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/config"
)

// math processes data by applying mathematic operations.
//
// This processor supports the object handling pattern.
type procMath struct {
	process
	Options procMathOptions `json:"options"`
}

type procMathOptions struct {
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
func (p procMath) String() string {
	return toString(p)
}

// Closes resources opened by the processor.
func (p procMath) Close(context.Context) error {
	return nil
}

// Batch processes one or more capsules with the processor. Conditions are
// optionally applied to the data to enable processing.
func (p procMath) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	return batchApply(ctx, capsules, p, p.Condition)
}

// Apply processes a capsule with the processor.
func (p procMath) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	// error early if required options are missing
	if p.Options.Operation == "" {
		return capsule, fmt.Errorf("process: math: options %+v: %v", p.Options, errMissingRequiredOptions)
	}

	// only supports JSON, error early if there are no keys
	if p.Key == "" && p.SetKey == "" {
		return capsule, fmt.Errorf("process: math: key %s set_key %s: %v", p.Key, p.SetKey, errInvalidDataPattern)
	}

	var value float64
	result := capsule.Get(p.Key)
	for i, res := range result.Array() {
		if i == 0 {
			value = res.Float()
			continue
		}

		switch p.Options.Operation {
		case "add":
			value += res.Float()
		case "subtract":
			value -= res.Float()
		case "multiply":
			value *= res.Float()
		case "divide":
			value /= res.Float()
		}
	}

	if err := capsule.Set(p.SetKey, value); err != nil {
		return capsule, fmt.Errorf("process: math: %v", err)
	}

	return capsule, nil
}
