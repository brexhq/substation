package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
)

// DeleteInvalidSettings is returned when the Copy processor is configured with invalid Input and Output settings.
const DeleteInvalidSettings = errors.Error("DeleteInvalidSettings")

/*
Delete processes encapsulated data by deleting JSON keys. The processor supports these patterns:
	JSON:
	  	{"foo":"bar","baz":"qux"} >>> {"foo":"bar"}

The processor uses this Jsonnet configuration:
	{
		type: 'delete',
		settings: {
			input_key: 'delete',
		}
	}
*/
type Delete struct {
	Condition condition.OperatorConfig `json:"condition"`
	InputKey  string                   `json:"input_key"`
}

// ApplyBatch processes a slice of encapsulated data with the Delete processor. Conditions are optionally applied to the data to enable processing.
func (p Delete) ApplyBatch(ctx context.Context, caps []config.Capsule) ([]config.Capsule, error) {
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

// Apply processes encapsulated data with the Delete processor.
func (p Delete) Apply(ctx context.Context, cap config.Capsule) (config.Capsule, error) {
	// only supports JSON, error early if there are no keys
	if p.InputKey == "" {
		return cap, fmt.Errorf("applicator settings %+v: %w", p, ProcessorInvalidSettings)
	}

	cap.Delete(p.InputKey)
	return cap, nil
}
