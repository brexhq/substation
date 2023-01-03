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

// errCaseInvalid is returned when the Case processor is configured with
// an invalid case.
const errCaseInvalid = errors.Error("invalid case")

// case processes data by modifying letter case (https://en.wikipedia.org/wiki/LetterprocCase).
//
// This processor supports the data and object handling patterns.
type procCase struct {
	process
	Options procCaseOptions `json:"options"`
}

type procCaseOptions struct {
	// Type is the case formatting that is applied.
	//
	// Must be one of:
	//
	// - upper
	//
	// - lower
	//
	// - snake
	Type string `json:"type"`
}

// String returns the processor settings as an object.
func (p procCase) String() string {
	return toString(p)
}

// Closes resources opened by the processor.
func (p procCase) Close(context.Context) error {
	return nil
}

// Batch processes one or more capsules with the processor. Conditions are
// optionally applied to the data to enable processing.
func (p procCase) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	return batchApply(ctx, capsules, p, p.Condition)
}

// Apply processes a capsule with the processor.
func (p procCase) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	// error early if required options are missing
	if p.Options.Type == "" {
		return capsule, fmt.Errorf("process: case: options %+v: %v", p.Options, errMissingRequiredOptions)
	}

	// JSON processing
	if p.Key != "" && p.SetKey != "" {
		result := capsule.Get(p.Key).String()

		var value string
		switch p.Options.Type {
		case "upper":
			value = strings.ToUpper(result)
		case "lower":
			value = strings.ToLower(result)
		case "snake":
			value = strcase.ToSnake(result)
		default:
			return capsule, fmt.Errorf("process: case: case %s: %v", p.Options.Type, errCaseInvalid)
		}

		if err := capsule.Set(p.SetKey, value); err != nil {
			return capsule, fmt.Errorf("process: case: %v", err)
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
			return capsule, fmt.Errorf("process: case: case %s: %v", p.Options.Type, errCaseInvalid)
		}

		capsule.SetData(value)
		return capsule, nil
	}

	return capsule, fmt.Errorf("process: case: key %s set_key %s: %v", p.Key, p.SetKey, errInvalidDataPattern)
}
