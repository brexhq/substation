package transform

import (
	"fmt"
	"strconv"

	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
)

type numberMathConfig struct {
	Object iconfig.Object `json:"object"`
}

func (c *numberMathConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *numberMathConfig) Validate() error {
	if c.Object.Key == "" && c.Object.SetKey != "" {
		return fmt.Errorf("object_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.Object.Key != "" && c.Object.SetKey == "" {
		return fmt.Errorf("object_set_key: %v", errors.ErrMissingRequiredOption)
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
