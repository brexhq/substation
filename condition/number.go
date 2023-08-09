package condition

import (
	"context"
	"fmt"
	"strconv"

	"golang.org/x/exp/slices"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
)

// number evaluates data using Types from the standard library's number package.
//
// This inspector supports the data and object handling patterns.
type inspNumber struct {
	condition
	Options inspNumberOptions `json:"options"`
}

type inspNumberOptions struct {
	// Type is the string evaluation Type used during inspection.
	//
	// Must be one of:
	//
	// - equals
	//
	// - greater_than
	//
	// - less_than
	//
	// - bitwise_and
	Type string `json:"type"`
	// Value is the length that is used for comparison during inspection.
	Value int64 `json:"value"`
}

// Creates a new number inspector.
func newInspNumber(_ context.Context, cfg config.Config) (c inspNumber, err error) {
	if err = config.Decode(cfg.Settings, &c); err != nil {
		return inspNumber{}, err
	}

	//  validate option.type
	if !slices.Contains(
		[]string{
			"equals",
			"greater_than",
			"less_than",
			"bitwise_and",
		},
		c.Options.Type) {
		return inspNumber{}, fmt.Errorf("condition: number: type %q: %v", c.Options.Type, errors.ErrInvalidOption)
	}

	return c, nil
}

func (c inspNumber) String() string {
	return toString(c)
}

// Inspect evaluates encapsulated data with the number inspector.
func (c inspNumber) Inspect(ctx context.Context, capsule config.Capsule) (output bool, err error) {
	var check int64
	if c.Key == "" {
		check, err = strconv.ParseInt(string(capsule.Data()), 10, 64)
		if err != nil {
			return false, fmt.Errorf("condition: number: invalid data processing value: %v", err)
		}
	} else {
		check = capsule.Get(c.Key).Int()
	}

	var matched bool
	switch c.Options.Type {
	case "equals":
		if check == c.Options.Value {
			matched = true
		}
	case "greater_than":
		if check > c.Options.Value {
			matched = true
		}
	case "less_than":
		if check < c.Options.Value {
			matched = true
		}
	case "bitwise_and":
		if check&c.Options.Value != 0 {
			matched = true
		}
	default:
		return false, fmt.Errorf("condition: number: type %s: %v", c.Options.Type, errors.ErrInvalidOption)
	}

	if c.Negate {
		return !matched, nil
	}

	return matched, nil
}
