package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
)

/*
InsertOptions contains custom options for the Insert processor:
	value:
		the value to insert
*/
type InsertOptions struct {
	Value interface{} `json:"value"`
}

/*
Insert processes encapsulated data by inserting a value into a JSON object. The processor supports these patterns:
	JSON:
		{"foo":"bar"} >>> {"foo":"bar","baz":"qux"}

The processor uses this Jsonnet configuration:
	{
		type: 'insert',
		settings: {
			options: {
				value: 'qux',
			},
			output_key: 'baz',
		},
	}
*/
type Insert struct {
	Options   InsertOptions            `json:"options"`
	Condition condition.OperatorConfig `json:"condition"`
	OutputKey string                   `json:"output_key"`
}

// ApplyBatch processes a slice of encapsulated data with the Insert processor. Conditions are optionally applied to the data to enable processing.
func (p Insert) ApplyBatch(ctx context.Context, caps []config.Capsule) ([]config.Capsule, error) {
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

// Apply processes encapsulated data with the Insert processor.
func (p Insert) Apply(ctx context.Context, cap config.Capsule) (config.Capsule, error) {
	// only supports JSON, error early if there are no keys
	if p.OutputKey == "" {
		return cap, fmt.Errorf("applicator settings %+v: %w", p, ProcessorInvalidSettings)
	}

	cap.Set(p.OutputKey, p.Options.Value)
	return cap, nil
}
