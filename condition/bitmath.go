package condition

import (
	"context"
	"fmt"
	"strconv"

	"golang.org/x/exp/slices"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
)

// bitmath evaluates data using bitwith math operations.
//
// This inspector supports the data and object handling patterns.
type inspBitmath struct {
	condition
	Options inspBitmathOptions `json:"options"`
}

type inspBitmathOptions struct {
	// Type is the string evaluation Type used during inspection.
	//
	// Must be one of:
	//
	// - and
	//
	// - or
	//
	// - not
	//
	// - xor
	Type string `json:"type"`
	// Value is the length that is used for comparison during inspection.
	Value int64 `json:"value"`
}

// Creates a new bitmath inspector.
func newInspBitmath(_ context.Context, cfg config.Config) (c inspBitmath, err error) {
	if err = config.Decode(cfg.Settings, &c); err != nil {
		return inspBitmath{}, err
	}

	//  validate option.type
	if !slices.Contains(
		[]string{
			"and",
			"or",
			"not",
			"xor",
		},
		c.Options.Type) {
		return inspBitmath{}, fmt.Errorf("condition: bitmath: type %q: %v", c.Options.Type, errors.ErrInvalidOption)
	}

	return c, nil
}

func (c inspBitmath) String() string {
	return toString(c)
}

// Inspect evaluates encapsulated data with the bitmath inspector.
func (c inspBitmath) Inspect(ctx context.Context, capsule config.Capsule) (output bool, err error) {
	var check int64
	if c.Key == "" {
		check, err = strconv.ParseInt(string(capsule.Data()), 10, 64)
		if err != nil {
			return false, fmt.Errorf("condition: bitmath: invalid data processing value: %v", err)
		}
	} else {
		check = capsule.Get(c.Key).Int()
	}

	var matched bool
	switch c.Options.Type {
	case "and":
		if check&c.Options.Value != 0 {
			matched = true
		}
	case "or":
		if check|c.Options.Value != 0 {
			matched = true
		}
	case "not":
		if ^check != 0 {
			matched = true
		}
	case "xor":
		if check^c.Options.Value != 0 {
			matched = true
		}
	default:
		return false, fmt.Errorf("condition: bitmath: type %s: %v", c.Options.Type, errors.ErrInvalidOption)
	}

	if c.Negate {
		return !matched, nil
	}

	return matched, nil
}
