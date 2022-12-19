package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/config"
)

type concat struct {
	process
	Options concatOptions `json:"options"`
}

type concatOptions struct {
	// Separator is the string that separates the values to be concatenated.
	Separator string `json:"separator"`
}

// Close closes resources opened by the Concat processor.
func (p concat) Close(context.Context) error {
	return nil
}

func (p concat) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	capsules, err := conditionalApply(ctx, capsules, p.Condition, p)
	if err != nil {
		return nil, fmt.Errorf("process capture: %v", err)
	}

	return capsules, nil
}

// Apply processes encapsulated data with the Concat processor.
func (p concat) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
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
