package condition

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/regexp"
)

// regExp evaluates data using a regular expression.
//
// This inspector supports the data and object handling patterns.
type regExp struct {
	condition
	Options regExpOptions `json:"options"`
}

type regExpOptions struct {
	// Expression is the regular expression used during inspection.
	Expression string `json:"expression"`
}

// Inspect evaluates encapsulated data with the regExp inspector.
func (c regExp) Inspect(ctx context.Context, capsule config.Capsule) (output bool, err error) {
	re, err := regexp.Compile(c.Options.Expression)
	if err != nil {
		return false, fmt.Errorf("condition regexp: %v", err)
	}

	var matched bool
	if c.Key == "" {
		matched = re.Match(capsule.Data())
	} else {
		res := capsule.Get(c.Key).String()
		matched = re.MatchString(res)
	}

	if c.Negate {
		return !matched, nil
	}

	return matched, nil
}
