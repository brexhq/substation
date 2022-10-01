package condition

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/regexp"
)

/*
RegExp evaluates data using a regular expression. This inspector uses a regexp cache provided by internal/regexp.

The inspector has these settings:
	Expression:
		the regular expression to use during inspection
	Key (optional):
		the JSON key-value to retrieve for inspection
	Negate (optional):
		if set to true, then the inspection is negated (i.e., true becomes false, false becomes true)
		defaults to false

The inspector supports these patterns:
	JSON:
		{"foo":"bar"} == ^bar
	data:
		bar == ^bar

When loaded with a factory, the inspector uses this JSON configuration:
	{
		"type": "regexp",
		"settings": {
			"expression": "^bar"
		},
	}
*/
type RegExp struct {
	Expression string `json:"expression"`
	Key        string `json:"key"`
	Negate     bool   `json:"negate"`
}

// Inspect evaluates encapsulated data with the RegExp inspector.
func (c RegExp) Inspect(ctx context.Context, cap config.Capsule) (output bool, err error) {
	re, err := regexp.Compile(c.Expression)
	if err != nil {
		return false, fmt.Errorf("condition regexp: %v", err)
	}

	var matched bool
	if c.Key == "" {
		matched = re.Match(cap.Data())
	} else {
		res := cap.Get(c.Key).String()
		matched = re.MatchString(res)
	}

	if c.Negate {
		return !matched, nil
	}

	return matched, nil
}
