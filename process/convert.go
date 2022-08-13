package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
)

/*
ConvertOptions contains custom options for the Convert processor:
	Type:
		the type that the value should be converted to
		must be one of:
			bool (boolean)
			int (integer)
			float
			uint (unsigned integer)
			string
*/
type ConvertOptions struct {
	Type string `json:"type"`
}

/*
Convert processes encapsulated data by converting values between types (e.g., string to integer, integer to float). The processor supports these patterns:
	JSON:
		{"convert":"true"} >>> {"convert":true}
		{"convert":"-123"} >>> {"convert":-123}
		{"convert":123} >>> {"convert":"123"}

The processor uses this Jsonnet configuration:
	{
		type: 'convert',
		settings: {
			options: {
				type: 'bool',
			},
			input_key: 'convert',
			output_key: 'convert',
		},
	}
*/
type Convert struct {
	Options   ConvertOptions           `json:"options"`
	Condition condition.OperatorConfig `json:"condition"`
	InputKey  string                   `json:"input_key"`
	OutputKey string                   `json:"output_key"`
}

// ApplyBatch processes a slice of encapsulated data with the Convert processor. Conditions are optionally applied to the data to enable processing.
func (p Convert) ApplyBatch(ctx context.Context, caps []config.Capsule) ([]config.Capsule, error) {
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

// Apply processes encapsulated data with the Convert processor.
func (p Convert) Apply(ctx context.Context, cap config.Capsule) (config.Capsule, error) {
	// error early if required options are missing
	if p.Options.Type == "" {
		return cap, fmt.Errorf("applicator settings %+v: %w", p, ProcessorInvalidSettings)
	}

	// only supports JSON, error early if there are no keys
	if p.InputKey != "" && p.OutputKey != "" {
		result := cap.Get(p.InputKey)

		switch p.Options.Type {
		case "bool":
			cap.Set(p.OutputKey, result.Bool())
		case "int":
			cap.Set(p.OutputKey, result.Int())
		case "float":
			cap.Set(p.OutputKey, result.Float())
		case "uint":
			cap.Set(p.OutputKey, result.Uint())
		case "string":
			cap.Set(p.OutputKey, result.String())
		}

		return cap, nil
	}

	return cap, fmt.Errorf("applicator settings %+v: %w", p, ProcessorInvalidSettings)
}
