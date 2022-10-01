package condition

import (
	"context"
	"fmt"
	"unicode/utf8"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
)

// lengthInvalidFunction is returned when the Length inspector is configured with an invalid function.
const lengthInvalidFunction = errors.Error("lengthInvalidFunction")

/*
Length evaluates data using len functions. This inspector supports evaluating byte and rune (character) length of strings. If a JSON array is input, then the length is evaluated against the number of elements in the array.

The inspector has these settings:
	Function:
		the length evaluation function used during inspection
		must be one of:
			equals
			greaterthan
			lessthan
	Value:
		the length value used during inspection
	Type (optional):
		the length type used during inpsection
		must be one of:
			byte (number of bytes)
			rune (number of characters)
		defaults to byte
	Key (optional):
		the JSON key-value to retrieve for inspection
	Negate (optional):
		if set to true, then the inspection is negated (i.e., true becomes false, false becomes true)
		defaults to false

The inspector supports these patterns:
	JSON:
		{"foo":"bar"} == 3
		{"foo":["bar","baz","qux"]} == 3
	data:
		bar == 3

When loaded with a factory, the inspector uses this JSON configuration:
	{
		"type": "length",
		"settings": {
			"function": "equals",
			"value": 3
		}
	}
*/
type Length struct {
	Function string `json:"function"`
	Value    int    `json:"value"`
	Type     string `json:"type"`
	Key      string `json:"key"`
	Negate   bool   `json:"negate"`
}

// Inspect evaluates encapsulated data with the Length inspector.
func (c Length) Inspect(ctx context.Context, cap config.Capsule) (output bool, err error) {
	var check string
	if c.Key == "" {
		check = string(cap.Data())
	} else {
		result := cap.Get(c.Key)
		if result.IsArray() {
			return c.match(len(result.Array()))
		}

		check = result.String()
	}

	var length int
	switch c.Type {
	case "byte":
		length = len(check)
	case "rune":
		length = utf8.RuneCountInString(check)
	default:
		length = len(check)
	}

	return c.match(length)
}

func (c Length) match(length int) (bool, error) {
	var matched bool
	switch c.Function {
	case "equals":
		if length == c.Value {
			matched = true
		}
	case "greaterthan":
		if length > c.Value {
			matched = true
		}
	case "lessthan":
		if length < c.Value {
			matched = true
		}
	default:
		return false, fmt.Errorf("condition length: function %s: %v", c.Function, lengthInvalidFunction)
	}

	if c.Negate {
		return !matched, nil
	}

	return matched, nil
}
