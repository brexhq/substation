package condition

import (
	"context"
	"fmt"
	"unicode/utf8"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
)

// errLengthInvalidType is returned when the length inspector is configured with an invalid Type.
const errLengthInvalidType = errors.Error("invalid Type")

// length evaluates data using len Types.
//
// This inspector supports the data and object handling patterns. If the input is an array, then the number of elements in the array is inspected.
type length struct {
	condition
	Options lengthOptions `json:"options"`
}

type lengthOptions struct {
	// Type determines the length evaluation Type used during inspection.
	//
	// Must be one of:
	//	- equals
	//	- greater_than
	//	- less_than
	Type string `json:"type"`
	// Value is the length that is used for comparison during inspection.
	Value int `json:"value"`
	// Measurement controls how the length is measured. The inspector automatically
	// assigns measurement for objects when the key is an array.
	//
	// Must be one of:
	//	- byte: number of bytes
	// 	- rune: number of characters
	//
	// This is optional and defaults to byte.
	Measurement string `json:"measurement"`
}

// Inspect evaluates encapsulated data with the length inspector.
func (c length) Inspect(ctx context.Context, capsule config.Capsule) (output bool, err error) {
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

func (c length) match(length int) (bool, error) {
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
		return false, fmt.Errorf("condition length: Type %s: %v", c.Options.Type, errLengthInvalidType)
	}

	if c.Negate {
		return !matched, nil
	}

	return matched, nil
}
