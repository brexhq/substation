package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
)

/*
Insert processes data by inserting a value into a JSON object. The processor supports these patterns:

	JSON:
		{"foo":"bar"} >>> {"foo":"bar","baz":"qux"}

When loaded with a factory, the processor uses this JSON configuration:

	{
		"type": "insert",
		"settings": {
			"options": {
				"value": "qux"
			},
			"output_key": "baz"
		}
	}
*/
type Insert struct {
	Options   InsertOptions    `json:"options"`
	Condition condition.Config `json:"condition"`
	OutputKey string           `json:"output_key"`
}

/*
InsertOptions contains custom options for the Insert processor:

	value:
		value to insert
*/
type InsertOptions struct {
	Value interface{} `json:"value"`
}

// Close closes resources opened by the Insert processor.
func (p Insert) Close(context.Context) error {
	return nil
}

// ApplyBatch processes a slice of encapsulated data with the Insert processor. Conditions are optionally applied to the data to enable processing.
func (p Insert) ApplyBatch(ctx context.Context, capsules []config.Capsule) ([]config.Capsule, error) {
	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, fmt.Errorf("process insert: %v", err)
	}

	capsules, err = conditionallyApplyBatch(ctx, capsules, op, p)
	if err != nil {
		return nil, fmt.Errorf("process insert: %v", err)
	}

	return capsules, nil
}

// Apply processes encapsulated data with the Insert processor.
func (p Insert) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	// only supports JSON, error early if there are no keys
	if p.OutputKey == "" {
		return capsule, fmt.Errorf("process insert: outputkey %s: %v", p.OutputKey, errInvalidDataPattern)
	}

	if err := capsule.Set(p.OutputKey, p.Options.Value); err != nil {
		return capsule, fmt.Errorf("process insert: %v", err)
	}

	return capsule, nil
}
