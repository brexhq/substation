package process

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/iancoleman/strcase"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
)

// errCaseInvalid is returned when the Case processor is configured with an invalid case.
const errCaseInvalid = errors.Error("invalid case")

type letterCase struct {
	process
	Options letterCaseOptions `json:"options"`
}

type letterCaseOptions struct {
	// Type is the case formatting that is applied.
	//
	// Must be one of:
	//	- uppercase
	//	- lowercase
	//	- snake
	Type string `json:"type"`
}

// Close closes resources opened by the Case processor.
func (p letterCase) Close(context.Context) error {
	return nil
}

func (p letterCase) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	capsules, err := conditionalApply(ctx, capsules, p.Condition, p)
	if err != nil {
		return nil, fmt.Errorf("process capture: %v", err)
	}

	return capsules, nil
}

// Apply processes encapsulated data with the Case processor.
func (p letterCase) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	// error early if required options are missing
	if p.Options.Type == "" {
		return capsule, fmt.Errorf("process case: options %+v: %v", p.Options, errMissingRequiredOptions)
	}

	// JSON processing
	if p.Key != "" && p.SetKey != "" {
		result := capsule.Get(p.Key).String()

		var value string
		switch p.Options.Type {
		case "uppercase":
			value = strings.ToUpper(result)
		case "lowercase":
			value = strings.ToLower(result)
		case "snake":
			value = strcase.ToSnake(result)
		default:
			return capsule, fmt.Errorf("process case: case %s: %v", p.Options.Type, errCaseInvalid)
		}

		if err := capsule.Set(p.SetKey, value); err != nil {
			return capsule, fmt.Errorf("process case: %v", err)
		}

		return capsule, nil
	}

	// data processing
	if p.Key == "" && p.SetKey == "" {
		var value []byte
		switch p.Options.Type {
		case "upper":
			value = bytes.ToUpper(capsule.Data())
		case "lower":
			value = bytes.ToLower(capsule.Data())
		default:
			return capsule, fmt.Errorf("process case: case %s: %v", p.Options.Type, errCaseInvalid)
		}

		capsule.SetData(value)
		return capsule, nil
	}

	return capsule, fmt.Errorf("process case: inputkey %s outputkey %s: %v", p.Key, p.SetKey, errInvalidDataPattern)
}
