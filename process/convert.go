package process

import (
	"context"
	"fmt"

	"golang.org/x/exp/slices"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
)

// convert processes data by changing its type (e.g., bool, int, string).
//
// This processor supports the object handling pattern.
type procConvert struct {
	process
	Options procConvertOptions `json:"options"`
}

type procConvertOptions struct {
	// Type is the target conversion type.
	//
	// Must be one of:
	//	- bool (boolean)
	//	- int (integer)
	//	- float
	//	- uint (unsigned integer)
	//	- string
	Type string `json:"type"`
}

// Create a new convert processor.
func newProcConvert(ctx context.Context, cfg config.Config) (p procConvert, err error) {
	if err = config.Decode(cfg.Settings, &p); err != nil {
		return procConvert{}, err
	}

	p.operator, err = condition.NewOperator(ctx, p.Condition)
	if err != nil {
		return procConvert{}, err
	}

	//  validate option.type
	if !slices.Contains(
		[]string{
			"bool",
			"int",
			"float",
			"uint",
			"string",
		},
		p.Options.Type) {
		return procConvert{}, fmt.Errorf("process: convert: type %q: %v", p.Options.Type, errors.ErrInvalidOption)
	}

	return p, nil
}

// String returns the processor settings as an object.
func (p procConvert) String() string {
	return toString(p)
}

// Closes resources opened by the processor.
func (p procConvert) Close(context.Context) error {
	return nil
}

// Batch processes one or more capsules with the processor. Conditions are
// optionally applied to the data to enable processing.
func (p procConvert) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	return batchApply(ctx, capsules, p, p.operator)
}

// Apply processes a capsule with the processor.
func (p procConvert) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	// only supports JSON, error early if there are no keys
	if p.Key != "" && p.SetKey != "" {
		result := capsule.Get(p.Key)

		var value interface{}
		switch p.Options.Type {
		case "bool":
			value = result.Bool()
		case "int":
			value = result.Int()
		case "float":
			value = result.Float()
		case "uint":
			value = result.Uint()
		case "string":
			value = result.String()
		}

		if err := capsule.Set(p.SetKey, value); err != nil {
			return capsule, fmt.Errorf("process: convert: %v", err)
		}

		return capsule, nil
	}

	return capsule, fmt.Errorf("process: convert: key %s set_key %s: %v", p.Key, p.SetKey, errInvalidDataPattern)
}
