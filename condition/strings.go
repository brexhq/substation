package condition

import (
	"strings"

	"github.com/brexhq/substation/internal/json"
)

// Strings implements the Inspector interface for evaluating data using string functions. More information is available in the README.
type Strings struct {
	Key        string `mapstructure:"key"`
	Expression string `mapstructure:"expression"`
	Function   string `mapstructure:"function"`
	Negate     bool   `mapstructure:"negate"`
}

// Inspect evaluates the data using string functions.
func (c Strings) Inspect(data []byte) (output bool, err error) {
	var check string
	if c.Key == "" {
		check = string(data)
	} else {
		check = json.Get(data, c.Key).String()
	}

	var matched bool
	switch f := c.Function; f {
	case "equals":
		if check == c.Expression {
			matched = true
		}
	case "contains":
		matched = strings.Contains(check, c.Expression)
	case "endswith":
		matched = strings.HasSuffix(check, c.Expression)
	case "startswith":
		matched = strings.HasPrefix(check, c.Expression)
	}

	if c.Negate {
		return !matched, nil
	}

	return matched, nil
}
