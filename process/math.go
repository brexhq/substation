package process

import (
	"context"
	"fmt"

	"golang.org/x/exp/slices"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
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

// Create a new math processor.
func newProcMath(cfg config.Config) (p procMath, err error) {
	if err = config.Decode(cfg.Settings, &p); err != nil {
		return procMath{}, err
	}

	p.operator, err = condition.NewOperator(p.Condition)
	if err != nil {
		return procMath{}, err
	}

	//  validate option.operation
	if !slices.Contains(
		[]string{
			"add",
			"subtract",
			"multiply",
			"divide",
		},
		p.Options.Operation) {
		return procMath{}, fmt.Errorf("process: math: operation %q: %v", p.Options.Operation, errors.ErrInvalidOption)
	}

	// only supports JSON, fail if there are no keys
	if p.Key == "" && p.SetKey == "" {
		return procMath{}, fmt.Errorf("process: math: key %s set_key %s: %v", p.Key, p.SetKey, errInvalidDataPattern)
	}

	return p, nil
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
	return batchApply(ctx, capsules, p, p.operator)
}

// Apply processes a capsule with the processor.
func (p procMath) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
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
