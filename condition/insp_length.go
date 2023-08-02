package condition

import (
	"context"
	"encoding/json"
	"fmt"
	"unicode/utf8"

	"golang.org/x/exp/slices"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
	mess "github.com/brexhq/substation/message"
)

type inspLengthConfig struct {
	// Key is the message key used during inspection.
	Key string `json:"key"`
	// Negate is a boolean that negates the inspection result.
	Negate bool `json:"negate"`
	// Type determines the length evaluation Type used during inspection.
	//
	// Must be one of:
	//
	// - equals
	//
	// - greater_than
	//
	// - less_than
	Type string `json:"type"`
	// Value is the length that is used for comparison during inspection.
	Value int `json:"value"`
	// Measurement controls how the length is measured. The inspector automatically
	// assigns measurement for objects when the key is an array.
	//
	// Must be one of:
	//
	// - byte: number of bytes
	//
	// - rune: number of characters
	//
	// This is optional and defaults to byte.
	Measurement string `json:"measurement"`
}

type inspLength struct {
	conf inspLengthConfig
}

func newInspLength(_ context.Context, cfg config.Config) (*inspLength, error) {
	conf := inspLengthConfig{}
	if err := config.Decode(cfg.Settings, &conf); err != nil {
		return nil, err
	}

	// Validate required options.
	if conf.Type == "" {
		return nil, fmt.Errorf("condition: insp_length: type: %v", errors.ErrMissingRequiredOption)
	}

	if !slices.Contains(
		[]string{
			"equals",
			"greater_than",
			"less_than",
		},
		conf.Type) {
		return nil, fmt.Errorf("condition: insp_length: type %q: %v", conf.Type, errors.ErrInvalidOption)
	}

	if conf.Measurement == "" {
		conf.Measurement = "byte"
	}

	if !slices.Contains(
		[]string{
			"byte",
			"rune",
		},
		conf.Measurement) {
		return nil, fmt.Errorf("condition: insp_length: measurement %q: %v", conf.Measurement, errors.ErrInvalidOption)
	}

	sink := inspLength{
		conf: conf,
	}

	return &sink, nil
}

func (c *inspLength) String() string {
	b, _ := json.Marshal(c.conf)
	return string(b)
}

func (c *inspLength) Inspect(ctx context.Context, message *mess.Message) (output bool, err error) {
	if message.IsControl() {
		return false, nil
	}

	var check string
	if c.conf.Key == "" {
		check = string(message.Data())
	} else {
		result := message.Get(c.conf.Key)
		if result.IsArray() {
			return c.match(len(result.Array()))
		}

		check = result.String()
	}

	var length int
	switch c.conf.Measurement {
	case "byte":
		length = len(check)
	case "rune":
		length = utf8.RuneCountInString(check)
	default:
		length = len(check)
	}

	return c.match(length)
}

func (c *inspLength) match(length int) (bool, error) {
	var matched bool
	switch c.conf.Type {
	case "equals":
		if length == c.conf.Value {
			matched = true
		}
	case "greater_than":
		if length > c.conf.Value {
			matched = true
		}
	case "less_than":
		if length < c.conf.Value {
			matched = true
		}
	}

	if c.conf.Negate {
		return !matched, nil
	}

	return matched, nil
}
