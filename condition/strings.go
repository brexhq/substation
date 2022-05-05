package condition

import (
	"strings"

	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/json"
)

// StringsInvalidFunction is returned when the Strings inspector is configured with an invalid function.
const StringsInvalidFunction = errors.Error("StringsInvalidFunction")

/*
Strings evaluates data using string functions. This inspector uses the standard library's strings package.

The inspector has these settings:
	Key (optional):
		the JSON key-value to retrieve for inspection
	Expression:
		the substring expression to use during inspection
	Function:
		the string evaluation function to use during inspection
		must be one of:
			equals
			contains
			endswith
			startswith
	Negate (optional):
		if set to true, then the inspection is negated (i.e., true becomes false, false becomes true)
		defaults to false

The inspector supports these patterns:
	json:
		{"foo":"bar"} == bar
	data:
		bar == bar

The inspector uses this Jsonnet configuration:
	{
		type: 'strings',
		settings: {
			key: 'foo',
			expression: 'bar',
			function: 'endswith',
		},
	}
*/
type Strings struct {
	Key        string `mapstructure:"key"`
	Expression string `mapstructure:"expression"`
	Function   string `mapstructure:"function"`
	Negate     bool   `mapstructure:"negate"`
}

// Inspect evaluates data with the Strings inspector.
func (c Strings) Inspect(data []byte) (output bool, err error) {
	var check string
	if c.Key == "" {
		check = string(data)
	} else {
		check = json.Get(data, c.Key).String()
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
		return false, StringsInvalidFunction
	}

	if c.Negate {
		return !matched, nil
	}

	return matched, nil
}
