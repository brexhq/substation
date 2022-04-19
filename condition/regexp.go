package condition

import (
	"fmt"

	"github.com/brexhq/substation/internal/json"
	"github.com/brexhq/substation/internal/regexp"
)

// RegExp implements the Inspector interface for evaluating data with a regular expression. More information is available in the README.
type RegExp struct {
	Key        string `mapstructure:"key"`
	Expression string `mapstructure:"expression"`
	Negate     bool   `mapstructure:"negate"`
}

// Inspect evaluates the data with a user-defined regular expression.
func (c RegExp) Inspect(data []byte) (output bool, err error) {
	var check string
	if c.Key == "" {
		check = string(data)
	} else {
		check = json.Get(data, c.Key).String()
	}

	re, err := regexp.Compile(c.Expression)
	if err != nil {
		return false, fmt.Errorf("err RegExp condition failed to compile regexp %s: %v", c.Expression, err)
	}

	matched := re.MatchString(check)
	if c.Negate {
		return !matched, nil
	}

	return matched, nil
}
