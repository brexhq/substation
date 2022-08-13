package process

import (
	"context"
	"fmt"
	"strings"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
)

/*
Copy processes encapsulated data by copying it. The processor supports these patterns:
	JSON:
	  	{"hello":"world"} >>> {"hello":"world","goodbye":"world"}
	from JSON:
  		{"hello":"world"} >>> world
	to JSON:
  		world >>> {"hello":"world"}

The processor uses this Jsonnet configuration:
	{
		type: 'copy',
		settings: {
			input_key: 'hello',
			output_key: 'goodbye',
		},
	}
*/
type Copy struct {
	Condition condition.OperatorConfig `json:"condition"`
	InputKey  string                   `json:"input_key"`
	OutputKey string                   `json:"output_key"`
}

// ApplyBatch processes a slice of encapsulated data with the Copy processor. Conditions are optionally applied to the data to enable processing.
func (p Copy) ApplyBatch(ctx context.Context, caps []config.Capsule) ([]config.Capsule, error) {
	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, fmt.Errorf("applybatch settings %+v: %w", p, err)
	}

	caps, err = conditionallyApplyBatch(ctx, caps, op, p)
	if err != nil {
		return nil, fmt.Errorf("applybatch settings %+v: %w", p, err)
	}

	return caps, nil
}

// Apply processes encapsulated data with the Copy processor.
func (p Copy) Apply(ctx context.Context, cap config.Capsule) (config.Capsule, error) {
	// JSON processing
	if p.InputKey != "" && p.OutputKey != "" {
		cap.Set(p.OutputKey, cap.Get(p.InputKey))
		return cap, nil
	}

	// from JSON processing
	if p.InputKey != "" && p.OutputKey == "" {
		res := cap.Get(p.InputKey).String()

		if strings.HasPrefix(p.InputKey, "__metadata") {
			cap.SetMetadata([]byte(res))
			return cap, nil
		}

		cap.SetData([]byte(res))
		return cap, nil
	}

	// to JSON processing
	if p.InputKey == "" && p.OutputKey != "" {
		if strings.HasPrefix(p.OutputKey, "__metadata") {
			cap.Set(p.OutputKey, cap.GetMetadata())
			return cap, nil
		}

		cap.Set(p.OutputKey, cap.GetData())
		return cap, nil
	}

	return cap, nil
}
