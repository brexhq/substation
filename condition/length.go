package condition

import (
	"fmt"
	"unicode/utf8"

	"github.com/brexhq/substation/config"
)

/*
Length evaluates data using len functions. This inspector supports evaluating byte and rune (character) length of strings. If a JSON array is input, then the length is evaluated against the number of elements in the array.

The inspector has these settings:
	Key (optional):
		the JSON key-value to retrieve for inspection
	Value:
		the length value used during inspection
	Type:
		the length type used during inpsection
		must be one of:
			byte (number of bytes)
			rune (number of characters)
		defaults to byte
	Function:
		the length evaluation function used during inspection
		must be one of:
			equals
			greaterthan
			lessthan
	Negate (optional):
		if set to true, then the inspection is negated (i.e., true becomes false, false becomes true)
		defaults to false

The inspector supports these patterns:
	JSON:
		{"foo":"bar"} == 3
		{"foo":["bar","baz","qux"]} == 3
	data:
		bar == 3

The inspector uses this Jsonnet configuration:
	{
		type: 'length',
		settings: {
			key: 'foo',
			value: 3,
			'function': 'equals',
		},
	}
*/
type Length struct {
	Key      string `json:"key"`
	Value    int    `json:"value"`
	Type     string `json:"type"`
	Function string `json:"function"`
	Negate   bool   `json:"negate"`
}

// Inspect evaluates data with the Length inspector.
func (c Length) Inspect(cap config.Capsule) (output bool, err error) {
	var check string
	if c.Key == "" {
		check = string(cap.GetData())
	} else {
		v := cap.Get(c.Key)
		if v.IsArray() {
			return c.match(len(v.Array()))
		}

		check = string(v.String())
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
		return false, fmt.Errorf("inspector settings %+v: %w", c, InspectorInvalidSettings)
	}

	if c.Negate {
		return !matched, nil
	}

	return matched, nil
}
