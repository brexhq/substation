package process

import (
	"bytes"
	"context"
	"strings"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/internal/json"
)

/*
SplitOptions contain custom options settings for this processor.

Separator: the substring that values are split from.
*/
type SplitOptions struct {
	Separator string `mapstructure:"separator"`
}

// Split implements the Byter and Channeler interfaces and Split bytes by a separator string. More information is available in the README.
type Split struct {
	Condition condition.OperatorConfig `mapstructure:"condition"`
	Input     Input                    `mapstructure:"input"`
	Output    Output                   `mapstructure:"output"`
	Options   SplitOptions             `mapstructure:"options"`
}

// Channel processes a data channel of bytes with this processor. Conditions can be optionally applied on the channel data to enable processing.
func (p Split) Channel(ctx context.Context, ch <-chan []byte) (<-chan []byte, error) {
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

		if p.Input.Key == "" && p.Output.Key == "" {
			processed := splitBytes(data, []byte(p.Options.Separator))
			for _, p := range processed {
				array = append(array, p)
			}
		} else {
			processed, err := p.Byte(ctx, data)
			if err != nil {
				return nil, err
			}
			array = append(array, processed)
		}
	}

	output := make(chan []byte, len(array))
	for _, x := range array {
		output <- x
	}
	close(output)
	return output, nil
}

// Byte processes a byte slice with this processor.
func (p Split) Byte(ctx context.Context, data []byte) ([]byte, error) {
	value := json.Get(data, p.Input.Key)

	if !value.IsArray() {
		var innerArray []string
		v := value.String()
		o := splitString(v, p.Options.Separator)
		for _, p := range o {
			innerArray = append(innerArray, string(p))
		}

		return json.Set(data, p.Output.Key, innerArray)
	}

	var outerArray [][]string
	for _, v := range value.Array() {
		var innerArray []string
		value := v.String()
		o := splitString(value, p.Options.Separator)
		for _, p := range o {
			innerArray = append(innerArray, string(p))
		}
		outerArray = append(outerArray, innerArray)
	}

	return json.Set(data, p.Output.Key, outerArray)
}

func splitBytes(v []byte, s []byte) [][]byte {
	return bytes.Split(v, s)
}

func splitString(v string, s string) []string {
	return strings.Split(v, s)
}
