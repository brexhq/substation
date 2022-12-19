package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/regexp"
)

// capture processes data by capturing values using regular expressions.
//
// This processor supports the data and object handling patterns.
type capture struct {
	process
	Options captureOptions `json:"options"`
}

type captureOptions struct {
	// Expression is the regular expression used to capture values.
	Expression string `json:"expression"`
	// Type determines which regular expression function is applied using the Expression.
	//
	// Must be one of:
	//	- find: applies the Find(String)?Submatch function
	//	- find_all: applies the FindAll(String)?Submatch function (see count)
	//	- named_group: applies the Find(String)?Submatch function and stores values as objects using subexpressions
	Type string `json:"type"`
	// Count manages the number of repeated capture groups.
	//
	// This is optional and defaults to match all capture groups.
	Count int `json:"count"`
}

// Close closes resources opened by the Capture processor.
func (p capture) Close(context.Context) error {
	return nil
}

func (p capture) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	capsules, err := conditionalApply(ctx, capsules, p.Condition, p)
	if err != nil {
		return nil, fmt.Errorf("process capture: %v", err)
	}

	return capsules, nil
}

// Apply processes encapsulated data with the Capture processor.
func (p capture) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	// error early if required options are missing
	if p.Options.Expression == "" || p.Options.Type == "" {
		return capsule, fmt.Errorf("process capture: options %+v: %v", p.Options, errMissingRequiredOptions)
	}

	re, err := regexp.Compile(p.Options.Expression)
	if err != nil {
		return capsule, fmt.Errorf("process capture: %v", err)
	}

	if p.Options.Count == 0 {
		p.Options.Count = -1
	}

	// JSON processing
	if p.Key != "" && p.SetKey != "" {
		result := capsule.Get(p.Key).String()

		switch p.Options.Type {
		case "find":
			match := re.FindStringSubmatch(result)
			if err := capsule.Set(p.SetKey, p.getStringMatch(match)); err != nil {
				return capsule, fmt.Errorf("process capture: %v", err)
			}

			return capsule, nil
		case "find_all":
			var matches []interface{}

			subs := re.FindAllStringSubmatch(result, p.Options.Count)
			for _, s := range subs {
				m := p.getStringMatch(s)
				matches = append(matches, m)
			}

			if err := capsule.Set(p.SetKey, matches); err != nil {
				return capsule, fmt.Errorf("process capture: %v", err)
			}

			return capsule, nil
		}
	}

	// data processing
	if p.Key == "" && p.SetKey == "" {
		switch p.Options.Type {
		case "find":
			match := re.FindSubmatch(capsule.Data())
			capsule.SetData(match[1])

			return capsule, nil
		case "named_group":
			newCapsule := config.NewCapsule()

			names := re.SubexpNames()
			matches := re.FindSubmatch(capsule.Data())
			for i, m := range matches {
				if i == 0 {
					continue
				}

				if err := newCapsule.Set(names[i], m); err != nil {
					return capsule, fmt.Errorf("process capture: %v", err)
				}
			}

			return newCapsule, nil
		}
	}

	return capsule, fmt.Errorf("process capture: inputkey %s outputkey %s: %v", p.Key, p.SetKey, errInvalidDataPattern)
}

func (p capture) getStringMatch(match []string) string {
	if len(match) > 1 {
		return match[len(match)-1]
	}

	return ""
}
