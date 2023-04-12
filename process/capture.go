package process

import (
	"context"
	"fmt"
	goregexp "regexp"

	"golang.org/x/exp/slices"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/regexp"
)

// capture processes data by capturing values using regular expressions.
//
// This processor supports the data and object handling patterns.
type procCapture struct {
	process
	Options procCaptureOptions `json:"options"`

	re *goregexp.Regexp
}

type procCaptureOptions struct {
	// Expression is the regular expression used to capture values.
	Expression string `json:"expression"`
	// Type determines which regular expression function is applied using
	// the Expression.
	//
	// Must be one of:
	//
	// - find: applies the Find(String)?Submatch function
	//
	// - find_all: applies the FindAll(String)?Submatch function (see count)
	//
	// - named_group: applies the Find(String)?Submatch function and stores
	// values as objects using subexpressions
	Type string `json:"type"`
	// Count manages the number of repeated capture groups.
	//
	// This is optional and defaults to match all capture groups.
	Count int `json:"count"`
}

// Create a new capture processor.
func newProcCapture(cfg config.Config) (p procCapture, err error) {
	if err = config.Decode(cfg.Settings, &p); err != nil {
		return procCapture{}, err
	}

	p.operator, err = condition.NewOperator(p.Condition)
	if err != nil {
		return procCapture{}, err
	}

	//  validate option.type
	if !slices.Contains(
		[]string{
			"find",
			"find_all",
			"named_group",
		},
		p.Options.Type) {
		return procCapture{}, fmt.Errorf("process: capture: type %q: %v", p.Options.Type, errors.ErrInvalidOptionInput)
	}

	// fail if required options are missing
	if p.Options.Expression == "" {
		return procCapture{}, fmt.Errorf("process: capture: option \"expression\": %v", errors.ErrMissingRequiredOption)
	}

	if _, err = regexp.Compile(p.Options.Expression); err != nil {
		return procCapture{}, fmt.Errorf("process: capture: %v", err)
	}

	if p.Options.Count == 0 {
		p.Options.Count = -1
	}

	p.re, err = regexp.Compile(p.Options.Expression)
	if err != nil {
		return procCapture{}, fmt.Errorf("process: capture: %v", err)
	}

	return p, nil
}

// String returns the processor settings as an object.
func (p procCapture) String() string {
	return toString(p)
}

// Closes resources opened by the processor.
func (p procCapture) Close(context.Context) error {
	return nil
}

// Batch processes one or more capsules with the processor. Conditions are
// optionally applied to the data to enable processing.
func (p procCapture) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	return batchApply(ctx, capsules, p, p.operator)
}

// Apply processes a capsule with the processor.
func (p procCapture) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	// JSON processing
	if p.Key != "" && p.SetKey != "" {
		result := capsule.Get(p.Key).String()

		switch p.Options.Type {
		case "find":
			match := p.re.FindStringSubmatch(result)
			if err := capsule.Set(p.SetKey, p.getStringMatch(match)); err != nil {
				return capsule, fmt.Errorf("process: capture: %v", err)
			}

			return capsule, nil
		case "find_all":
			var matches []interface{}

			subs := p.re.FindAllStringSubmatch(result, p.Options.Count)
			for _, s := range subs {
				m := p.getStringMatch(s)
				matches = append(matches, m)
			}

			if err := capsule.Set(p.SetKey, matches); err != nil {
				return capsule, fmt.Errorf("process: capture: %v", err)
			}

			return capsule, nil
		case "named_group":
			names := p.re.SubexpNames()
			matches := p.re.FindStringSubmatch(result)
			for i, m := range matches {
				if i == 0 {
					continue
				}

				// if the same key is used multiple times, then this will correctly
				// set multiple named groups into that key.
				//
				// if set_key is "foo" and the first group returns {"bar":"baz"}, then
				// the output is {"foo":{"bar":"baz"}}. if the second group returns
				// {"qux":"quux"} then the output is {"foo":{"bar":"baz","qux":"quux"}}.
				setKey := p.SetKey + "." + names[i]
				if err := capsule.Set(setKey, m); err != nil {
					return capsule, fmt.Errorf("process: capture: %v", err)
				}
			}

			return capsule, nil
		}
	}

	// data processing
	if p.Key == "" && p.SetKey == "" {
		switch p.Options.Type {
		case "find":
			match := p.re.FindSubmatch(capsule.Data())
			capsule.SetData(match[1])

			return capsule, nil
		case "named_group":
			newCapsule := config.NewCapsule()

			names := p.re.SubexpNames()
			matches := p.re.FindSubmatch(capsule.Data())
			for i, m := range matches {
				if i == 0 {
					continue
				}

				if err := newCapsule.Set(names[i], m); err != nil {
					return capsule, fmt.Errorf("process: capture: %v", err)
				}
			}

			return newCapsule, nil
		}
	}

	return capsule, fmt.Errorf("process: capture: key %s set_key %s: %v", p.Key, p.SetKey, errInvalidDataPattern)
}

func (p procCapture) getStringMatch(match []string) string {
	if len(match) > 1 {
		return match[len(match)-1]
	}

	return ""
}
