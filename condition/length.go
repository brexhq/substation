package condition

import (
	"context"
	"fmt"
	"unicode/utf8"

	"golang.org/x/exp/slices"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
)

// length evaluates data using len Types.
//
// This inspector supports the data and object handling patterns. If the input is an array, then the number of elements in the array is inspected.
type inspLength struct {
	condition
	Options inspLengthOptions `json:"options"`
}

type inspLengthOptions struct {
	// Type determines the length evaluation Type used during inspection.
	//
	// Must be one of:
	//
	// - equals
	//
	// - greater_than
	//
	// - less_than
	Type string `json:"type"`
	// Value is the length that is used for comparison during inspection.
	Value int `json:"value"`
	// Measurement controls how the length is measured. The inspector automatically
	// assigns measurement for objects when the key is an array.
	//
	// Must be one of:
	//
	// - byte: number of bytes
	//
	// - rune: number of characters
	//
	// This is optional and defaults to byte.
	Measurement string `json:"measurement"`
}

// Creates a new length inspector.
func newInspLength(cfg config.Config) (c inspLength, err error) {
	err = config.Decode(cfg.Settings, &c)
	if err != nil {
		return inspLength{}, err
	}

	//  validate option.type
	if !slices.Contains(
		[]string{
			"equals",
			"greater_than",
			"less_than",
		},
		c.Options.Type) {
		return inspLength{}, fmt.Errorf("condition: length: type %q invalid: %w", c.Options.Type, errors.ErrMissingRequiredOptions)
	}

	if c.Options.Measurement == "" {
		c.Options.Measurement = "byte"
	}
	//  validate option.type
	if !slices.Contains(
		[]string{
			"byte",
			"rune",
		},
		c.Options.Measurement) {
		return inspLength{}, fmt.Errorf("condition: length: measurement %q invalid: %w", c.Options.Measurement, errors.ErrMissingRequiredOptions)
	}

	return c, nil
}

func (c inspLength) String() string {
	return toString(c)
}

// Inspect evaluates encapsulated data with the length inspector.
func (c inspLength) Inspect(ctx context.Context, capsule config.Capsule) (output bool, err error) {
	var check string
	if c.Key == "" {
		check = string(capsule.Data())
	} else {
		result := capsule.Get(c.Key)
		if result.IsArray() {
			return c.match(len(result.Array()))
		}

		check = result.String()
	}

	var length int
	switch c.Options.Measurement {
	case "byte":
		length = len(check)
	case "rune":
		length = utf8.RuneCountInString(check)
	default:
		length = len(check)
	}

	return c.match(length)
}

func (c inspLength) match(length int) (bool, error) {
	var matched bool
	switch c.Options.Type {
	case "equals":
		if length == c.Options.Value {
			matched = true
		}
	case "greater_than":
		if length > c.Options.Value {
			matched = true
		}
	case "less_than":
		if length < c.Options.Value {
			matched = true
		}
	default:
		return false, fmt.Errorf("condition: length: type %s: %v", c.Options.Type, errors.ErrInvalidType)
	}

	if c.Negate {
		return !matched, nil
	}

	return matched, nil
}
