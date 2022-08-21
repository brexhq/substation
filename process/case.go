package process

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/iancoleman/strcase"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
)

/*
Case processes data by changing the case of a string or byte slice. The processor supports these patterns:
	JSON:
		{"case":"foo"} >>> {"case":"FOO"}
	data:
		foo >>> FOO

When loaded with a factory, the processor uses this JSON configuration:
	{
		"type": "case",
		"settings": {
			"options": {
				"case": "upper"
			},
			"input_key": "case",
			"output_key": "case"
		}
	}
*/
type Case struct {
	Options   CaseOptions      `json:"options"`
	Condition condition.Config `json:"condition"`
	InputKey  string           `json:"input_key"`
	OutputKey string           `json:"output_key"`
}

/*
CaseOptions contains custom options for the Case processor:
	Case:
		the case to convert the string or byte to
		must be one of:
			upper
			lower
			snake (strings only)
*/
type CaseOptions struct {
	Case string `json:"case"`
}

// ApplyBatch processes a slice of encapsulated data with the Case processor. Conditions are optionally applied to the data to enable processing.
func (p Case) ApplyBatch(ctx context.Context, caps []config.Capsule) ([]config.Capsule, error) {
	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, fmt.Errorf("applybatch settings %+v: %v", p, err)
	}

	caps, err = conditionallyApplyBatch(ctx, caps, op, p)
	if err != nil {
		return nil, fmt.Errorf("applybatch settings %+v: %v", p, err)
	}

	return caps, nil
}

// Apply processes encapsulated data with the Case processor.
func (p Case) Apply(ctx context.Context, cap config.Capsule) (config.Capsule, error) {
	// error early if required options are missing
	if p.Options.Case == "" {
		return cap, fmt.Errorf("apply settings %+v: %w", p, ProcessorInvalidSettings)
	}

	// JSON processing
	if p.InputKey != "" && p.OutputKey != "" {
		res := cap.Get(p.InputKey).String()
		switch p.Options.Case {
		case "upper":
			cap.Set(p.OutputKey, strings.ToUpper(res))
		case "lower":
			cap.Set(p.OutputKey, strings.ToLower(res))
		case "snake":
			cap.Set(p.OutputKey, strcase.ToSnake(res))
		}

		return cap, nil
	}

	// data processing
	if p.InputKey == "" && p.OutputKey == "" {
		switch p.Options.Case {
		case "upper":
			cap.SetData(bytes.ToUpper(cap.GetData()))
		case "lower":
			cap.SetData(bytes.ToLower(cap.GetData()))
		}

		return cap, nil
	}

	return cap, fmt.Errorf("apply settings %+v: %w", p, ProcessorInvalidSettings)
}
