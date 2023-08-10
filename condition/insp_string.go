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

type inspStringConfig struct {
	// Key is the message key used during inspection.
	Key string `json:"key"`
	// Negate is a boolean that negates the inspection result.
	Negate bool `json:"negate"`
	// Type determines the string evaluation method used during inspection.
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
	// Value is a string used during inspection.
	String string `json:"string"`
}

// string evaluates data using Types from the standard library's string package.
//
// This inspector supports the data and object handling patterns.
type inspString struct {
	conf inspStringConfig
}

// Creates a new string inspector.
func newInspString(_ context.Context, cfg config.Config) (*inspString, error) {
	conf := inspStringConfig{}
	if err := _config.Decode(cfg.Settings, &conf); err != nil {
		return nil, err
	}

	// Validate required options.
	if conf.Type == "" {
		return nil, fmt.Errorf("condition: insp_string: type: %v", errors.ErrMissingRequiredOption)
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
		return nil, fmt.Errorf("condition: insp_string: type %q: %v", conf.Type, errors.ErrInvalidOption)
	}

	insp := inspString{
		conf: conf,
	}

	return &insp, nil
}

func (c *inspString) String() string {
	b, _ := json.Marshal(c.conf)
	return string(b)
}

func (c *inspString) Inspect(ctx context.Context, message *mess.Message) (output bool, err error) {
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
	switch c.conf.Type {
	case "equals":
		if check == c.conf.String {
			matched = true
		}
	case "contains":
		matched = strings.Contains(check, c.conf.String)
	case "starts_with":
		matched = strings.HasPrefix(check, c.conf.String)
	case "ends_with":
		matched = strings.HasSuffix(check, c.conf.String)
	case "greater_than":
		matched = strings.Compare(check, c.conf.String) > 0
	case "less_than":
		matched = strings.Compare(check, c.conf.String) < 0
	}

	if c.conf.Negate {
		return !matched, nil
	}

	return matched, nil
}
