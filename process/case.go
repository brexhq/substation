package process

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/iancoleman/strcase"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
)

// errCaseInvalid is returned when the Case processor is configured with an invalid case.
const errCaseInvalid = errors.Error("invalid case")

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
		case to convert the string or byte to
		must be one of:
			upper
			lower
			snake (strings only)
*/
type CaseOptions struct {
	Case string `json:"case"`
}

// ApplyBatch processes a slice of encapsulated data with the Case processor. Conditions are optionally applied to the data to enable processing.
func (p Case) ApplyBatch(ctx context.Context, capsules []config.Capsule) ([]config.Capsule, error) {
	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, fmt.Errorf("process case: %v", err)
	}

	capsules, err = conditionallyApplyBatch(ctx, capsules, op, p)
	if err != nil {
		return nil, fmt.Errorf("process case: %v", err)
	}

	return capsules, nil
}

// Apply processes encapsulated data with the Case processor.
func (p Case) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	// error early if required options are missing
	if p.Options.Case == "" {
		return capsule, fmt.Errorf("process case: options %+v: %v", p.Options, errMissingRequiredOptions)
	}

	// JSON processing
	if p.InputKey != "" && p.OutputKey != "" {
		result := capsule.Get(p.InputKey).String()

		var value string
		switch p.Options.Case {
		case "upper":
			value = strings.ToUpper(result)
		case "lower":
			value = strings.ToLower(result)
		case "snake":
			value = strcase.ToSnake(result)
		default:
			return capsule, fmt.Errorf("process case: case %s: %v", p.Options.Case, errCaseInvalid)
		}

		if err := capsule.Set(p.OutputKey, value); err != nil {
			return capsule, fmt.Errorf("process case: %v", err)
		}

		return capsule, nil
	}

	// data processing
	if p.InputKey == "" && p.OutputKey == "" {
		var value []byte
		switch p.Options.Case {
		case "upper":
			value = bytes.ToUpper(capsule.Data())
		case "lower":
			value = bytes.ToLower(capsule.Data())
		default:
			return capsule, fmt.Errorf("process case: case %s: %v", p.Options.Case, errCaseInvalid)
		}

		capsule.SetData(value)
		return capsule, nil
	}

	return capsule, fmt.Errorf("process case: inputkey %s outputkey %s: %v", p.InputKey, p.OutputKey, errInvalidDataPattern)
}
