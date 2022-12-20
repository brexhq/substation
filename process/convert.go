package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/config"
)

// convert processes data by changing its type (e.g., bool, int, string).
//
// This processor supports the object handling pattern.
type _convert struct {
	process
	Options _convertOptions `json:"options"`
}

type _convertOptions struct {
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

// Close closes resources opened by the processor.
func (p _convert) Close(context.Context) error {
	return nil
}

// Batch processes one or more capsules with the processor. Conditions are
// optionally applied to the data to enable processing.
func (p _convert) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	return conditionalApply(ctx, capsules, p.Condition, p)
}

// Apply processes a capsule with the processor.
func (p _convert) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	// error early if required options are missing
	if p.Options.Type == "" {
		return capsule, fmt.Errorf("process convert: options %+v: %v", p.Options, errMissingRequiredOptions)
	}

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
			return capsule, fmt.Errorf("process convert: %v", err)
		}

		return capsule, nil
	}

	return capsule, fmt.Errorf("process convert: inputkey %s outputkey %s: %v", p.Key, p.SetKey, errInvalidDataPattern)
}
