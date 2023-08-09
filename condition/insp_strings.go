package condition

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"golang.org/x/exp/slices"

	"github.com/brexhq/substation/config"
	_config "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	mess "github.com/brexhq/substation/message"
)

type inspStringsConfig struct {
	// Key is the message key used during inspection.
	Key string `json:"key"`
	// Negate is a boolean that negates the inspection result.
	Negate bool `json:"negate"`
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

// strings evaluates data using Types from the standard library's strings package.
//
// This inspector supports the data and object handling patterns.
type inspStrings struct {
	conf inspStringsConfig
}

// Creates a new strings inspector.
func newInspStrings(_ context.Context, cfg config.Config) (*inspStrings, error) {
	conf := inspStringsConfig{}
	if err := _config.Decode(cfg.Settings, &conf); err != nil {
		return nil, err
	}

	// Validate required options.
	if conf.Type == "" {
		return nil, fmt.Errorf("condition: insp_strings: type: %v", errors.ErrMissingRequiredOption)
	}

	if !slices.Contains(
		[]string{
			"equals",
			"contains",
			"starts_with",
			"ends_with",
			"greater_than",
			"less_than",
		},
		conf.Type) {
		return nil, fmt.Errorf("condition: insp_strings: type %q: %v", conf.Type, errors.ErrInvalidOption)
	}

	insp := inspStrings{
		conf: conf,
	}

	return &insp, nil
}

func (c *inspStrings) String() string {
	b, _ := json.Marshal(c.conf)
	return string(b)
}

func (c *inspStrings) Inspect(ctx context.Context, message *mess.Message) (output bool, err error) {
	if message.IsControl() {
		return false, nil
	}

	var check string
	if c.conf.Key == "" {
		check = string(message.Data())
	} else {
		check = message.Get(c.conf.Key).String()
	}

	var matched bool
	switch s := c.conf.Type; s {
	case "equals":
		if check == c.conf.Expression {
			matched = true
		}
	case "contains":
		matched = strings.Contains(check, c.conf.Expression)
	case "starts_with":
		matched = strings.HasPrefix(check, c.conf.Expression)
	case "ends_with":
		matched = strings.HasSuffix(check, c.conf.Expression)
	case "greater_than":
		matched = strings.Compare(check, c.conf.Expression) > 0
	case "less_than":
		matched = strings.Compare(check, c.conf.Expression) < 0
	}

	if c.conf.Negate {
		return !matched, nil
	}

	return matched, nil
}
