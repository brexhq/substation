package process

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/iancoleman/strcase"
	"golang.org/x/exp/slices"

	"github.com/brexhq/substation/condition"
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

// Create a new case processor.
func newProcCase(cfg config.Config) (p procCase, err error) {
	if err = config.Decode(cfg.Settings, &p); err != nil {
		return procCase{}, err
	}

	p.operator, err = condition.NewOperator(p.Condition)
	if err != nil {
		return procCase{}, err
	}

	//  validate option.type
	if !slices.Contains(
		[]string{
			"upper",
			"lower",
			"snake",
		},
		p.Options.Type) {
		return procCase{}, fmt.Errorf("process: case: type %q: %v", p.Options, errors.ErrInvalidOptionInput)
	}

	// validate data processing pattern
	if (p.Key != "" && p.SetKey == "") ||
		(p.Key == "" && p.SetKey != "") {
		return procCase{}, fmt.Errorf("process: case: key %s set_key %s: %v", p.Key, p.SetKey, errInvalidDataPattern)
	}

	return p, nil
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
	return batchApply(ctx, capsules, p, p.operator)
}

// Apply processes a capsule with the processor.
func (p procCase) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
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
