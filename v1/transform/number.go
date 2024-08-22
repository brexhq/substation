package transform

import (
	"fmt"
	"strconv"

	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
)

// Use this config for any Number transform that only requires a single value.
type numberValConfig struct {
	Value float64 `json:"value"`

	ID     string         `json:"id"`
	Object iconfig.Object `json:"object"`
}

func (c *numberValConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

// 0.0 is a valid value and should not be checked.
func (c *numberValConfig) Validate() error {
	if c.Object.SourceKey == "" && c.Object.TargetKey != "" {
		return fmt.Errorf("object_source_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.Object.SourceKey != "" && c.Object.TargetKey == "" {
		return fmt.Errorf("object_target_key: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

type numberMathConfig struct {
	ID     string         `json:"id"`
	Object iconfig.Object `json:"object"`
}

func (c *numberMathConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *numberMathConfig) Validate() error {
	if c.Object.SourceKey == "" && c.Object.TargetKey != "" {
		return fmt.Errorf("object_source_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.Object.SourceKey != "" && c.Object.TargetKey == "" {
		return fmt.Errorf("object_target_key: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

// numberFloat64ToString addresses multiple issues with performing math
// operations on floats:
//
//   - Converts the float to a string without scientific notation: 1.1e+9 -> 1100000000
//
//   - Truncates the float to remove trailing zeros: 1.100000000 -> 1.1
//
//   - Removes the decimal point if it is a whole number: 1.0 -> 1
func numberFloat64ToString(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}
