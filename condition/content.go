package condition

import (
	"net/http"
	"strings"
)

// Content implements the Inspector interface for evaluating data by content type. More information is available in the README.
type Content struct {
	Type   string `mapstructure:"type"`
	Negate bool   `mapstructure:"negate"`
}

// Inspect evaluates the data by content type.
func (c Content) Inspect(data []byte) (output bool, err error) {
	var matched bool

	content := http.DetectContentType(data)
	if strings.Compare(content, c.Type) == 0 {
		matched = true
	} else {
		matched = false
	}

	if c.Negate {
		return !matched, nil
	}

	return matched, nil
}
