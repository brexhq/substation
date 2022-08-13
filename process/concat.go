package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
)

/*
ConcatOptions contains custom options for the Concat processor:
	Separator:
		the string that separates the concatenated values
*/
type ConcatOptions struct {
	Separator string `json:"separator"`
}

/*
Concat processes encapsulated data by concatenating multiple values together with a separator. The processor supports these patterns:
	JSON:
		{"concat":["foo","bar"]} >>> {"concat":"foo.bar"}

The processor uses this Jsonnet configuration:
	{
		type: 'concat',
		settings: {
			options: {
				separator: '.',
			},
			input_key: 'concat',
			output_key: 'concat',
		},
	}
*/
type Concat struct {
	Options   ConcatOptions            `json:"options"`
	Condition condition.OperatorConfig `json:"condition"`
	InputKey  string                   `json:"input_key"`
	OutputKey string                   `json:"output_key"`
}

// ApplyBatch processes a slice of encapsulated data with the Concat processor. Conditions are optionally applied to the data to enable processing.
func (p Concat) ApplyBatch(ctx context.Context, caps []config.Capsule) ([]config.Capsule, error) {
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

// Apply processes encapsulated data with the Concat processor.
func (p Concat) Apply(ctx context.Context, cap config.Capsule) (config.Capsule, error) {
	// error early if required options are missing
	if p.Options.Separator == "" {
		return cap, fmt.Errorf("applicator settings %+v: %w", p, ProcessorInvalidSettings)
	}

	// only supports JSON, error early if there are no keys
	if p.InputKey == "" && p.OutputKey == "" {
		return cap, fmt.Errorf("applicator settings %+v: %w", p, ProcessorInvalidSettings)
	}

	// data is processed by retrieving and iterating the
	// array (InputKey) containing string values and joining
	// each one with the separator string
	//
	// root:
	// 	{"concat":["foo","bar","baz"]}
	// concatenated:
	// 	{"concat:"foo.bar.baz"}
	var tmp string
	res := cap.Get(p.InputKey)
	for idx, val := range res.Array() {
		tmp += val.String()
		if idx != len(res.Array())-1 {
			tmp += p.Options.Separator
		}
	}

	cap.Set(p.OutputKey, tmp)
	return cap, nil
}
