package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/config"
)

// join processes data by concatenating values in an object array.
//
// This processor supports the object handling pattern.
type _join struct {
	process
	Options _joinOptions `json:"options"`
}

type _joinOptions struct {
	// Separator is the string that joins data from the array.
	Separator string `json:"separator"`
}

// Close closes resources opened by the processor.
func (p _join) Close(context.Context) error {
	return nil
}

// Batch processes one or more capsules with the processor. Conditions are
// optionally applied to the data to enable processing.
func (p _join) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	return conditionalApply(ctx, capsules, p.Condition, p)
}

// Apply processes encapsulated data with the processor.
func (p _join) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	// error early if required options are missing
	if p.Options.Separator == "" {
		return capsule, fmt.Errorf("process concat: options %+v: %v", p.Options, errMissingRequiredOptions)
	}

	// only supports JSON, error early if there are no keys
	if p.Key == "" && p.SetKey == "" {
		return capsule, fmt.Errorf("process concat: inputkey %s outputkey %s: %v", p.Key, p.SetKey, errInvalidDataPattern)
	}

	// data is processed by retrieving and iterating the
	// array (Key) containing string values and joining
	// each one with the separator string
	//
	// root:
	// 	{"concat":["foo","bar","baz"]}
	// concatenated:
	// 	{"concat:"foo.bar.baz"}
	var value string
	result := capsule.Get(p.Key)
	for i, res := range result.Array() {
		value += res.String()
		if i != len(result.Array())-1 {
			value += p.Options.Separator
		}
	}

	if err := capsule.Set(p.SetKey, value); err != nil {
		return capsule, fmt.Errorf("process dynamodb: %v", err)
	}

	return capsule, nil
}
