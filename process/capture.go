package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/internal/json"
	"github.com/brexhq/substation/internal/regexp"
)

/*
CaptureOptions contain custom options settings for this processor.

Expression: the capturing regular expression
Count (optional): the number of captures to return; defaults to 0, which returns all captures
*/
type CaptureOptions struct {
	Expression string `mapstructure:"expression"`
	Count      int    `mapstructure:"count"`
}

// Capture implements the Byter and Channeler interfaces and applies a capturing regular expression to data. More information is available in the README.
type Capture struct {
	Condition condition.OperatorConfig `mapstructure:"condition"`
	Input     Input                    `mapstructure:"input"`
	Output    Output                   `mapstructure:"output"`
	Options   CaptureOptions           `mapstructure:"options"`
}

// Channel processes a data channel of bytes with this processor. Conditions can be optionally applied on the channel data to enable processing.
func (p Capture) Channel(ctx context.Context, ch <-chan []byte) (<-chan []byte, error) {
	var array [][]byte

	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, err
	}

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

// Byte processes a byte slice with this processor
func (p Capture) Byte(ctx context.Context, data []byte) ([]byte, error) {
	value := json.Get(data, p.Input.Key)

	if !value.IsArray() {
		s := value.String()
		matches, err := p.capture(s)
		if err != nil {
			return nil, err
		}

		// if only one match, then set the element directly in the JSON key
		if len(matches) == 1 {
			return json.Set(data, p.Output.Key, matches[0])
		}

		return json.Set(data, p.Output.Key, matches)
	}

	// array needs to be able to hold strings and slices of strings
	// 	if it contains slices of strings, then the output should be additionally
	// 	processed by the flatten processor
	var array []interface{}
	for _, v := range value.Array() {
		s := v.String()
		matches, err := p.capture(s)
		if err != nil {
			return nil, err
		}

		// if only one ok, then append the element directly in the array
		if len(matches) == 1 {
			array = append(array, matches[0])
			continue
		}

		array = append(array, matches)
	}

	return json.Set(data, p.Output.Key, array)
}

func (p Capture) capture(v string) ([]string, error) {
	re, err := regexp.Compile(p.Options.Expression)
	if err != nil {
		return nil, fmt.Errorf("err Capture processor failed to compile regexp %s: %v", p.Options.Expression, err)
	}

	subs := re.FindAllStringSubmatch(v, p.Options.Count)
	var matches []string
	for _, s := range subs {
		m, _ := getMatch(s)

		matches = append(matches, m)
	}

	return matches, nil
}

func getMatch(m []string) (o string, l int) {
	if len(m) > 1 {
		o = m[len(m)-1]
	}

	return o, len(o)
}
