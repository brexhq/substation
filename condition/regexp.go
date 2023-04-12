package condition

import (
	"context"
	"fmt"
	"regexp"

	"github.com/brexhq/substation/config"
)

// regExp evaluates data using a regular expression.
//
// This inspector supports the data and object handling patterns.
type inspRegExp struct {
	condition
	Options inspRegExpOptions `json:"options"`

	re *regexp.Regexp
}

type inspRegExpOptions struct {
	// Expression is the regular expression used during inspection.
	Expression string `json:"expression"`
}

// Creates a new regexp inspector.
func newInspRegExp(cfg config.Config) (c inspRegExp, err error) {
	if err = config.Decode(cfg.Settings, &c); err != nil {
		return inspRegExp{}, err
	}

	c.re, err = regexp.Compile(c.Options.Expression)
	if err != nil {
		return inspRegExp{}, fmt.Errorf("condition: regexp: %v", err)
	}

	return c, nil
}

func (c inspRegExp) String() string {
	return toString(c)
}

// Inspect evaluates encapsulated data with the regExp inspector.
func (c inspRegExp) Inspect(ctx context.Context, capsule config.Capsule) (output bool, err error) {
	var matched bool
	if c.Key == "" {
		matched = c.re.Match(capsule.Data())
	} else {
		res := capsule.Get(c.Key).String()
		matched = c.re.MatchString(res)
	}

	if c.Negate {
		return !matched, nil
	}

	return matched, nil
}
