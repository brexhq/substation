package condition

import (
	"fmt"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/regexp"
)

/*
RegExp evaluates encapsulated data using a regular expression. This inspector uses a regexp cache provided by internal/regexp.

The inspector has these settings:
	Key (optional):
		the JSON key-value to retrieve for inspection
	Expression:
		the regular expression to use during inspection
	Negate (optional):
		if set to true, then the inspection is negated (i.e., true becomes false, false becomes true)
		defaults to false

The inspector supports these patterns:
	json:
		{"foo":"bar"} == ^bar
	data:
		bar == ^bar

The inspector uses this Jsonnet configuration:
	{
		type: 'regexp',
		settings: {
			key: 'foo',
			expression: '^bar',
		},
	}
*/
type RegExp struct {
	Key        string `json:"key"`
	Expression string `json:"expression"`
	Negate     bool   `json:"negate"`
}

// Inspect evaluates encapsulated data with the RegExp inspector.
func (c RegExp) Inspect(cap config.Capsule) (output bool, err error) {
	re, err := regexp.Compile(c.Expression)
	if err != nil {
		return false, fmt.Errorf("inspector settings %v: %v", c, err)
	}

	var matched bool
	if c.Key == "" {
		matched = re.Match(cap.GetData())
	} else {
		s := cap.Get(c.Key).String()
		matched = re.MatchString(s)
	}

	if c.Negate {
		return !matched, nil
	}

	return matched, nil
}
