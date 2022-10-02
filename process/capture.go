package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/regexp"
)

/*
Capture processes data by capturing values using regular expressions. The processor supports these patterns:
	JSON:
		{"capture":"foo@qux.com"} >>> {"capture":"foo"}
		{"capture":"foo@qux.com"} >>> {"capture":["f","o","o"]}
	data:
		foo@qux.com >>> foo
		bar quux >>> {"foo":"bar","qux":"quux"}

When loaded with a factory, the processor uses this JSON configuration:
	{
		"type": "capture",
		"settings": {
			"options": {
				"expression": "^([^@]*)@.*$",
				"function": "find"
			},
			"input_key": "capture",
			"output_key": "capture"
		}
	}
*/
type Capture struct {
	Options   CaptureOptions   `json:"options"`
	Condition condition.Config `json:"condition"`
	InputKey  string           `json:"input_key"`
	OutputKey string           `json:"output_key"`
}

/*
CaptureOptions contains custom options for the Capture processor:
	Expression:
		the regular expression used to capture values
	Function:
		the type of regular expression applied
		must be one of:
			find: applies the Find(String)?Submatch function
			find_all: applies the FindAll(String)?Submatch function (see count)
			named_group: applies the Find(String)?Submatch function and stores values as JSON using subexpressions
	Count (optional):
		used for repeating capture groups
		defaults to match all capture groups
*/
type CaptureOptions struct {
	Expression string `json:"expression"`
	Function   string `json:"function"`
	Count      int    `json:"count"`
}

// ApplyBatch processes a slice of encapsulated data with the Capture processor. Conditions are optionally applied to the data to enable processing.
func (p Capture) ApplyBatch(ctx context.Context, caps []config.Capsule) ([]config.Capsule, error) {
	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, fmt.Errorf("process capture: %v", err)
	}

	caps, err = conditionallyApplyBatch(ctx, caps, op, p)
	if err != nil {
		return nil, fmt.Errorf("process capture: %v", err)
	}

	return caps, nil
}

// Apply processes encapsulated data with the Capture processor.
func (p Capture) Apply(ctx context.Context, cap config.Capsule) (config.Capsule, error) {
	// error early if required options are missing
	if p.Options.Expression == "" || p.Options.Function == "" {
		return cap, fmt.Errorf("process capture: options %+v: %v", p.Options, errMissingRequiredOptions)
	}

	re, err := regexp.Compile(p.Options.Expression)
	if err != nil {
		return cap, fmt.Errorf("process capture: %v", err)
	}

	if p.Options.Count == 0 {
		p.Options.Count = -1
	}

	// JSON processing
	if p.InputKey != "" && p.OutputKey != "" {
		result := cap.Get(p.InputKey).String()

		switch p.Options.Function {
		case "find":
			match := re.FindStringSubmatch(result)
			if err := cap.Set(p.OutputKey, p.getStringMatch(match)); err != nil {
				return cap, fmt.Errorf("process capture: %v", err)
			}

			return cap, nil
		case "find_all":
			var matches []interface{}

			subs := re.FindAllStringSubmatch(result, p.Options.Count)
			for _, s := range subs {
				m := p.getStringMatch(s)
				matches = append(matches, m)
			}

			if err := cap.Set(p.OutputKey, matches); err != nil {
				return cap, fmt.Errorf("process capture: %v", err)
			}

			return cap, nil
		}
	}

	// data processing
	if p.InputKey == "" && p.OutputKey == "" {
		switch p.Options.Function {
		case "find":
			match := re.FindSubmatch(cap.Data())
			cap.SetData(match[1])

			return cap, nil
		case "named_group":
			newCap := config.NewCapsule()

			names := re.SubexpNames()
			matches := re.FindSubmatch(cap.Data())
			for i, m := range matches {
				if i == 0 {
					continue
				}

				if err := newCap.Set(names[i], m); err != nil {
					return cap, fmt.Errorf("process capture: %v", err)
				}
			}

			return newCap, nil
		}
	}

	return cap, fmt.Errorf("process capture: inputkey %s outputkey %s: %v", p.InputKey, p.OutputKey, errInvalidDataPattern)
}

func (p Capture) getStringMatch(match []string) string {
	if len(match) > 1 {
		return match[len(match)-1]
	}

	return ""
}
