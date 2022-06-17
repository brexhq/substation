package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/json"
	"github.com/brexhq/substation/internal/regexp"
)

// CaptureInvalidSettings is returned when the Capture processor is configured with invalid Input and Output settings.
const CaptureInvalidSettings = errors.Error("CaptureInvalidSettings")

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

/*
Capture processes data by capturing values using regular expressions. The processor supports these patterns:
	json:
		{"capture":"foo@qux.com"} >>> {"capture":"foo"}
		{"capture":"foo@qux.com"} >>> {"capture":["f","o","o"]}
	json array:
		{"capture":["foo@qux.com","bar@qux.com"]} >>> {"capture":["foo","bar"]}
		{"capture":["foo@qux.com","bar@qux.com"]} >>> {"capture":[["f","o","o"],["b","a","r"]]}
	data:
		foo@qux.com >>> foo
		bar quux >>> {"foo":"bar","qux":"quux"}

The processor uses this Jsonnet configuration:
	{
		type: 'capture',
		settings: {
			input_key: 'capture',
			output_key: 'capture',
			options: {
				expression: '^([^@]*)@.*$',
				_function: 'find',
			},
		},
	}
*/
type Capture struct {
	Condition condition.OperatorConfig `json:"condition"`
	InputKey  string                   `json:"input_key"`
	OutputKey string                   `json:"output_key"`
	Options   CaptureOptions           `json:"options"`
}

// Slice processes a slice of bytes with the Capture processor. Conditions are optionally applied on the bytes to enable processing.
func (p Capture) Slice(ctx context.Context, s [][]byte) ([][]byte, error) {
	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, fmt.Errorf("slicer settings %v: %v", p, err)
	}

	slice := NewSlice(&s)
	for _, data := range s {
		ok, err := op.Operate(data)
		if err != nil {
			return nil, fmt.Errorf("slicer settings %v: %v", p, err)
		}

		if !ok {
			slice = append(slice, data)
			continue
		}

		processed, err := p.Byte(ctx, data)
		if err != nil {
			return nil, fmt.Errorf("slicer: %v", err)
		}
		slice = append(slice, processed)
	}

	return slice, nil
}

// Byte processes bytes with the Capture processor.
func (p Capture) Byte(ctx context.Context, data []byte) ([]byte, error) {
	re, err := regexp.Compile(p.Options.Expression)
	if err != nil {
		return nil, fmt.Errorf("byter settings %v: %v", p, err)
	}

	if p.Options.Count == 0 {
		p.Options.Count = -1
	}

	// json processing
	if p.InputKey != "" && p.OutputKey != "" {
		value := json.Get(data, p.InputKey)

		if !value.IsArray() {
			if p.Options.Function == "find" {
				match := re.FindStringSubmatch(value.String())
				return json.Set(data, p.OutputKey, p.getStringMatch(match))
			}

			if p.Options.Function == "find_all" {
				var matches []interface{}

				subs := re.FindAllStringSubmatch(value.String(), p.Options.Count)
				for _, s := range subs {
					m := p.getStringMatch(s)
					matches = append(matches, m)
				}

				return json.Set(data, p.OutputKey, matches)
			}
		}

		// json array processing
		var array []interface{}
		for _, v := range value.Array() {
			var matches []interface{}

			if p.Options.Function == "find" {
				match := re.FindStringSubmatch(v.String())
				array = append(array, p.getStringMatch(match))
				continue
			}

			if p.Options.Function == "find_all" {
				subs := re.FindAllStringSubmatch(v.String(), p.Options.Count)
				for _, s := range subs {
					m := p.getStringMatch(s)
					matches = append(matches, m)
				}

				array = append(array, matches)
			}
		}

		return json.Set(data, p.OutputKey, array)
	}

	// data processing
	if p.InputKey == "" && p.OutputKey == "" {
		if p.Options.Function == "find" {
			match := re.FindSubmatch(data)
			return match[1], nil
		}

		if p.Options.Function == "named_group" {
			names := re.SubexpNames()

			var tmp []byte
			matches := re.FindSubmatch(data)
			for i, m := range matches {
				tmp, err = json.Set(tmp, names[i], m)
			}

			return tmp, nil
		}
	}

	return nil, fmt.Errorf("byter settings %v: %v", p, CaptureInvalidSettings)
}

func (p Capture) getStringMatch(match []string) string {
	if len(match) > 1 {
		return match[len(match)-1]
	}

	return ""
}
