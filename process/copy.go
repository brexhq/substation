package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
)

/*
Copy processes data by copying it. The processor supports these patterns:

		JSON:
		  	{"hello":"world"} >>> {"hello":"world","goodbye":"world"}
		from JSON:
	  		{"hello":"world"} >>> world
		to JSON:
	  		world >>> {"hello":"world"}

When loaded with a factory, the processor uses this JSON configuration:

	{
		"type": "copy",
		"settings": {
			"input_key": "hello",
			"output_key": "goodbye"
		}
	}
*/
type Copy struct {
	Condition condition.Config `json:"condition"`
	InputKey  string           `json:"input_key"`
	OutputKey string           `json:"output_key"`
}

// ApplyBatch processes a slice of encapsulated data with the Copy processor. Conditions are optionally applied to the data to enable processing.
func (p Copy) ApplyBatch(ctx context.Context, capsules []config.Capsule) ([]config.Capsule, error) {
	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, fmt.Errorf("process copy: %v", err)
	}

	capsules, err = conditionallyApplyBatch(ctx, capsules, op, p)
	if err != nil {
		return nil, fmt.Errorf("process copy: %v", err)
	}

	return capsules, nil
}

// Apply processes encapsulated data with the Copy processor.
func (p Copy) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	// JSON processing
	if p.InputKey != "" && p.OutputKey != "" {
		if err := capsule.Set(p.OutputKey, capsule.Get(p.InputKey)); err != nil {
			return capsule, fmt.Errorf("process copy: %v", err)
		}

		return capsule, nil
	}

	// from JSON processing
	if p.InputKey != "" && p.OutputKey == "" {
		result := capsule.Get(p.InputKey).String()

		capsule.SetData([]byte(result))
		return capsule, nil
	}

	// to JSON processing
	if p.InputKey == "" && p.OutputKey != "" {
		if err := capsule.Set(p.OutputKey, capsule.Data()); err != nil {
			return capsule, fmt.Errorf("process copy: %v", err)
		}

		return capsule, nil
	}

	return capsule, fmt.Errorf("process copy: inputkey %s outputkey %s: %w", p.InputKey, p.OutputKey, errInvalidDataPattern)
}
