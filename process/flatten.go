package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
)

/*
FlattenOptions contains custom options settings for the Flatten processor:
	Deep (optional):
		deeply flattens nested arrays
*/
type FlattenOptions struct {
	Deep bool `json:"deep"`
}

/*
Flatten processes encapsulated data by flattening JSON arrays. The processor supports these patterns:
	JSON:
		{"flatten":["foo",["bar"]]} >>> {"flatten":["foo","bar"]}

The processor uses this Jsonnet configuration:
	{
		type: 'flatten',
		settings: {
			input_key: 'flatten',
			output_key: 'flatten',
		},
	}
*/
type Flatten struct {
	Options   FlattenOptions           `json:"options"`
	Condition condition.OperatorConfig `json:"condition"`
	InputKey  string                   `json:"input_key"`
	OutputKey string                   `json:"output_key"`
}

// ApplyBatch processes a slice of encapsulated data with the Flatten processor. Conditions are optionally applied to the data to enable processing.
func (p Flatten) ApplyBatch(ctx context.Context, caps []config.Capsule) ([]config.Capsule, error) {
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

// Apply processes encapsulated data with the Flatten processor.
func (p Flatten) Apply(ctx context.Context, cap config.Capsule) (config.Capsule, error) {
	// only supports JSON, error early if there are no keys
	if p.InputKey == "" && p.OutputKey == "" {
		return cap, fmt.Errorf("applicator settings %+v: %w", p, ProcessorInvalidSettings)
	}

	if p.Options.Deep {
		cap.Set(p.OutputKey, cap.Get(p.InputKey+`|@flatten:{"deep":true}`))
	} else {
		cap.Set(p.OutputKey, cap.Get(p.InputKey+`|@flatten`))
	}

	return cap, nil
}
