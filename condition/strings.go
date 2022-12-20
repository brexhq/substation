package condition

import (
	"context"
	"fmt"
	"strings"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
)

// errStringsInvalidType is returned when the strings inspector is configured with an invalid Type.
const errStringsInvalidType = errors.Error("invalid Type")

// strings evaluates data using Types from the standard library's strings package.
//
// This inspector supports the data and object handling patterns.
type _strings struct {
	condition
	Options _stringsOptions `json:"options"`
}

type _stringsOptions struct {
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
	Type string `json:"type"`
	// Expression is a substring used during inspection.
	Expression string `json:"expression"`
}

func (c _strings) String() string {
	return inspectorToString(c)
}

// Inspect evaluates encapsulated data with the strings inspector.
func (c _strings) Inspect(ctx context.Context, capsule config.Capsule) (output bool, err error) {
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
	default:
		return false, fmt.Errorf("condition strings: Type %s: %v", c.Options.Type, errStringsInvalidType)
	}

	if c.Negate {
		return !matched, nil
	}

	return matched, nil
}
