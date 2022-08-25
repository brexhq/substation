package condition

import (
	"context"
	"fmt"
	"strings"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
)

// StringsInvalidFunction is returned when the Strings inspector is configured with an invalid function.
const StringsInvalidFunction = errors.Error("StringsInvalidFunction")

/*
Strings evaluates data using string functions. This inspector uses the standard library's strings package.

The inspector has these settings:
	Function:
		the string evaluation function to use during inspection
		must be one of:
			equals
			contains
			endswith
			startswith
	Expression:
		the substring expression to use during inspection
	Key (optional):
		the JSON key-value to retrieve for inspection
	Negate (optional):
		if set to true, then the inspection is negated (i.e., true becomes false, false becomes true)
		defaults to false

The inspector supports these patterns:
	JSON:
		{"foo":"bar"} == bar
	data:
		bar == bar

When loaded with a factory, the inspector uses this JSON configuration:
	{
		"type": "strings",
		"settings": {
			"function": "endswith",
			"expression": "bar"
		}
	}
*/
type Strings struct {
	Function   string `json:"function"`
	Expression string `json:"expression"`
	Key        string `json:"key"`
	Negate     bool   `json:"negate"`
}

// Inspect evaluates encapsulated data with the Strings inspector.
func (c Strings) Inspect(ctx context.Context, cap config.Capsule) (output bool, err error) {
	var check string
	if c.Key == "" {
		check = string(cap.GetData())
	} else {
		check = cap.Get(c.Key).String()
	}

	var matched bool
	switch s := c.Function; s {
	case "equals":
		if check == c.Expression {
			matched = true
		}
	case "contains":
		matched = strings.Contains(check, c.Expression)
	case "endswith":
		matched = strings.HasSuffix(check, c.Expression)
	case "startswith":
		matched = strings.HasPrefix(check, c.Expression)
	default:
		return false, fmt.Errorf("condition strings: function %s: %v", c.Function, StringsInvalidFunction)
	}

	if c.Negate {
		return !matched, nil
	}

	return matched, nil
}
