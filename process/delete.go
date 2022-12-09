package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
)

/*
Delete processes data by deleting JSON keys. The processor supports these patterns:

	JSON:
	  	{"foo":"bar","baz":"qux"} >>> {"foo":"bar"}

When loaded with a factory, the processor uses this JSON configuration:

	{
		"type": "delete",
		"settings": {
			"input_key": "delete"
		}
	}
*/
type Delete struct {
	Condition condition.Config `json:"condition"`
	InputKey  string           `json:"input_key"`
}

// Close closes resources opened by the Delete processor.
func (p Delete) Close(context.Context) error {
	return nil
}

// ApplyBatch processes a slice of encapsulated data with the Delete processor. Conditions are optionally applied to the data to enable processing.
func (p Delete) ApplyBatch(ctx context.Context, capsules []config.Capsule) ([]config.Capsule, error) {
	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, fmt.Errorf("process delete: %v", err)
	}

	capsules, err = conditionallyApplyBatch(ctx, capsules, op, p)
	if err != nil {
		return nil, fmt.Errorf("process delete: %v", err)
	}

	return capsules, nil
}

// Apply processes encapsulated data with the Delete processor.
func (p Delete) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	// only supports JSON, error early if there are no keys
	if p.InputKey == "" {
		return capsule, fmt.Errorf("process delete: inputkey %s: %v", p.InputKey, errInvalidDataPattern)
	}

	if err := capsule.Delete(p.InputKey); err != nil {
		return capsule, fmt.Errorf("process delete: %v", err)
	}

	return capsule, nil
}
