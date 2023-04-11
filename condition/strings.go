package condition

import (
	"context"
	"fmt"
	"strings"

	"golang.org/x/exp/slices"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
)

// strings evaluates data using Types from the standard library's strings package.
//
// This inspector supports the data and object handling patterns.
type inspStrings struct {
	condition
	Options inspStringsOptions `json:"options"`
}

type inspStringsOptions struct {
	// Type is the string evaluation Type used during inspection.
	//
	// Must be one of:
	//
	// - equals
	//
	// - contains
	//
	// - starts_with
	//
	// - ends_with
	//
	// - greater_than
	//
	// - less_than
	Type string `json:"type"`
	// Expression is a substring used during inspection.
	Expression string `json:"expression"`
}

// Creates a new strings inspector.
func newInspStrings(cfg config.Config) (c inspStrings, err error) {
	err = config.Decode(cfg.Settings, &c)
	if err != nil {
		return inspStrings{}, err
	}

	//  validate option.type
	if !slices.Contains(
		[]string{
			"equals",
			"contains",
			"starts_with",
			"ends_with",
		},
		c.Options.Type) {
		return inspStrings{}, fmt.Errorf("condition: strings: type %q invalid: %v", c.Options.Type, errors.ErrInvalidOptionInput)
	}

	// TODO(shellcromancer): should we check that the expression != "" if Type != "equals"?

	return c, nil
}

func (c inspStrings) String() string {
	return toString(c)
}

// Inspect evaluates encapsulated data with the strings inspector.
func (c inspStrings) Inspect(ctx context.Context, capsule config.Capsule) (output bool, err error) {
	var check string
	if c.Key == "" {
		check = string(capsule.Data())
	} else {
		check = capsule.Get(c.Key).String()
	}

	var matched bool
	switch s := c.Options.Type; s {
	case "equals":
		if check == c.Options.Expression {
			matched = true
		}
	case "contains":
		matched = strings.Contains(check, c.Options.Expression)
	case "starts_with":
		matched = strings.HasPrefix(check, c.Options.Expression)
	case "ends_with":
		matched = strings.HasSuffix(check, c.Options.Expression)
	case "greater_than":
		matched = strings.Compare(check, c.Options.Expression) > 0
	case "less_than":
		matched = strings.Compare(check, c.Options.Expression) < 0
	default:
		return false, fmt.Errorf("condition: strings: type %s: %v", c.Options.Type, errors.ErrInvalidType)
	}

	if c.Negate {
		return !matched, nil
	}

	return matched, nil
}
