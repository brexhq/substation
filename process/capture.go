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
	Expression string `mapstructure:"expression"`
	Function   string `mapstructure:"function"`
	Count      int    `mapstructure:"count"`
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
			input: {
				key: 'capture',
			},
			output: {
				key: 'capture',
			},
			options: {
				expression: '^([^@]*)@.*$',
				function: 'find',
			},
		},
	}
*/
type Capture struct {
	Condition condition.OperatorConfig `mapstructure:"condition"`
	Input     Input                    `mapstructure:"input"`
	Output    Output                   `mapstructure:"output"`
	Options   CaptureOptions           `mapstructure:"options"`
}

// Channel processes a data channel of byte slices with the Capture processor. Conditions are optionally applied on the channel data to enable processing.
func (p Capture) Channel(ctx context.Context, ch <-chan []byte) (<-chan []byte, error) {
	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, err
	}

	var array [][]byte
	for data := range ch {
		ok, err := op.Operate(data)
		if err != nil {
			return nil, err
		}

		if !ok {
			array = append(array, data)
			continue
		}

		processed, err := p.Byte(ctx, data)
		if err != nil {
			return nil, err
		}
		array = append(array, processed)
	}

	output := make(chan []byte, len(array))

	for _, x := range array {
		output <- x
	}
	close(output)
	return output, nil

}

// Byte processes a byte slice with the Capture processor.
func (p Capture) Byte(ctx context.Context, data []byte) ([]byte, error) {
	re, err := regexp.Compile(p.Options.Expression)
	if err != nil {
		return nil, fmt.Errorf("err Capture processor failed to compile regexp %s: %v", p.Options.Expression, err)
	}

	if p.Options.Count == 0 {
		p.Options.Count = -1
	}

	// json processing
	if p.Input.Key != "" && p.Output.Key != "" {
		value := json.Get(data, p.Input.Key)

		if !value.IsArray() {
			if p.Options.Function == "find" {
				match := re.FindStringSubmatch(value.String())
				return json.Set(data, p.Output.Key, match[1])
			}

			if p.Options.Function == "find_all" {
				var matches []interface{}

				subs := re.FindAllStringSubmatch(value.String(), p.Options.Count)
				for _, s := range subs {
					m := p.getStringMatch(s)
					matches = append(matches, m)
				}

				return json.Set(data, p.Output.Key, matches)
			}
		}

		// json array processing
		var array []interface{}
		for _, v := range value.Array() {
			var matches []interface{}

			if p.Options.Function == "find" {
				match := re.FindStringSubmatch(v.String())
				array = append(array, match[1])
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

		return json.Set(data, p.Output.Key, array)
	}

	// data processing
	if p.Input.Key == "" && p.Output.Key == "" {
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

	return nil, CaptureInvalidSettings
}

func (p Capture) getStringMatch(match []string) string {
	if len(match) > 1 {
		return match[len(match)-1]
	}

	return ""
}
