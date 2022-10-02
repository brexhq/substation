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
func (p Copy) ApplyBatch(ctx context.Context, caps []config.Capsule) ([]config.Capsule, error) {
	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, fmt.Errorf("process copy: %v", err)
	}

	caps, err = conditionallyApplyBatch(ctx, caps, op, p)
	if err != nil {
		return nil, fmt.Errorf("process copy: %v", err)
	}

	return caps, nil
}

// Apply processes encapsulated data with the Copy processor.
func (p Copy) Apply(ctx context.Context, cap config.Capsule) (config.Capsule, error) {
	// JSON processing
	if p.InputKey != "" && p.OutputKey != "" {
		if err := cap.Set(p.OutputKey, cap.Get(p.InputKey)); err != nil {
			return cap, fmt.Errorf("process copy: %v", err)
		}

		return cap, nil
	}

	// from JSON processing
	if p.InputKey != "" && p.OutputKey == "" {
		result := cap.Get(p.InputKey).String()

		cap.SetData([]byte(result))
		return cap, nil
	}

	// to JSON processing
	if p.InputKey == "" && p.OutputKey != "" {
		if err := cap.Set(p.OutputKey, cap.Data()); err != nil {
			return cap, fmt.Errorf("process copy: %v", err)
		}

		return cap, nil
	}

	return cap, fmt.Errorf("process copy: inputkey %s outputkey %s: %v", p.InputKey, p.OutputKey, errProcessorInvalidDataPattern)
}
