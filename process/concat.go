package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
)

/*
Concat processes data by concatenating multiple values together with a separator. The processor supports these patterns:
	JSON:
		{"concat":["foo","bar"]} >>> {"concat":"foo.bar"}

When loaded with a factory, the processor uses this JSON configuration:
	{
		"type": "concat",
		"settings": {
			"options": {
				"separator": "."
			},
			"input_key": "concat",
			"output_key": "concat"
		}
	}
*/
type Concat struct {
	Options   ConcatOptions    `json:"options"`
	Condition condition.Config `json:"condition"`
	InputKey  string           `json:"input_key"`
	OutputKey string           `json:"output_key"`
}

/*
ConcatOptions contains custom options for the Concat processor:
	Separator:
		the string that separates the concatenated values
*/
type ConcatOptions struct {
	Separator string `json:"separator"`
}

// ApplyBatch processes a slice of encapsulated data with the Concat processor. Conditions are optionally applied to the data to enable processing.
func (p Concat) ApplyBatch(ctx context.Context, caps []config.Capsule) ([]config.Capsule, error) {
	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, fmt.Errorf("process concat: %v", err)
	}

	caps, err = conditionallyApplyBatch(ctx, caps, op, p)
	if err != nil {
		return nil, fmt.Errorf("process concat: %v", err)
	}

	return caps, nil
}

// Apply processes encapsulated data with the Concat processor.
func (p Concat) Apply(ctx context.Context, cap config.Capsule) (config.Capsule, error) {
	// error early if required options are missing
	if p.Options.Separator == "" {
		return cap, fmt.Errorf("process concat: options %+v: %v", p.Options, errMissingRequiredOptions)
	}

	// only supports JSON, error early if there are no keys
	if p.InputKey == "" && p.OutputKey == "" {
		return cap, fmt.Errorf("process concat: inputkey %s outputkey %s: %v", p.InputKey, p.OutputKey, errInvalidDataPattern)
	}

	// data is processed by retrieving and iterating the
	// array (InputKey) containing string values and joining
	// each one with the separator string
	//
	// root:
	// 	{"concat":["foo","bar","baz"]}
	// concatenated:
	// 	{"concat:"foo.bar.baz"}
	var value string
	result := cap.Get(p.InputKey)
	for i, res := range result.Array() {
		value += res.String()
		if i != len(result.Array())-1 {
			value += p.Options.Separator
		}
	}

	if err := cap.Set(p.OutputKey, value); err != nil {
		return cap, fmt.Errorf("process dynamodb: %v", err)
	}

	return cap, nil
}
