package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/config"
)

/*
math processes data by applying mathematic operations. The processor supports these patterns:

	JSON:
		{"math":[1,3]} >>> {"math":4}

When loaded with a factory, the processor uses this JSON configuration:

	{
		"type": "math",
		"settings": {
			"options": {
				"operation": "add"
			},
			"input_key": "math",
			"output_key": "math"
		}
	}
*/
type math struct {
	process
	Options mathOptions `json:"options"`
}

type mathOptions struct {
	Operation string `json:"operation"`
}

// Close closes resources opened by the math processor.
func (p math) Close(context.Context) error {
	return nil
}

// ApplyBatch processes a slice of encapsulated data with the math processor. Conditions are optionally applied to the data to enable processing.
func (p math) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	capsules, err := conditionalApply(ctx, capsules, p.Condition, p)
	if err != nil {
		return nil, fmt.Errorf("process math: %v", err)
	}

	return capsules, nil
}

// Apply processes encapsulated data with the math processor.
func (p math) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
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
