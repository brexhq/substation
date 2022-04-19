package process

import (
	"context"
	"strings"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/internal/json"
)

/*
ReplaceOptions contain custom options settings for this processor.

Old: the substring to replace
New: the substring that replaces
Count: the number of replacements to make; defaults to 0, which replaces all substrings
*/
type ReplaceOptions struct {
	Old   string `mapstructure:"old"`
	New   string `mapstructure:"new"`
	Count int    `mapstructure:"count"`
}

// Replace implements the Byter and Channeler interfaces and replaces substrings in string values. More information is available in the README.
type Replace struct {
	Condition condition.OperatorConfig `mapstructure:"condition"`
	Input     Input                    `mapstructure:"input"`
	Output    Output                   `mapstructure:"output"`
	Options   ReplaceOptions           `mapstructure:"options"`
}

func _replace(v string, o ReplaceOptions) string {
	return strings.Replace(v, o.Old, o.New, o.Count)
}

// Channel processes a data channel of bytes with this processor. Conditions can be optionally applied on the channel data to enable processing.
func (p Replace) Channel(ctx context.Context, ch <-chan []byte) (<-chan []byte, error) {
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
func (p Replace) Byte(ctx context.Context, data []byte) ([]byte, error) {
	// default to replace all
	if p.Options.Count == 0 {
		p.Options.Count = -1
	}

	value := json.Get(data, p.Input.Key)

	if !value.IsArray() {
		s := value.String()
		o := p.replace(s)
		return json.Set(data, p.Output.Key, o)
	}

	var array []string
	for _, v := range value.Array() {
		s := v.String()
		o := p.replace(s)
		array = append(array, o)
	}

	return json.Set(data, p.Output.Key, array)
}

func (p Replace) replace(s string) string {
	return strings.Replace(s, p.Options.Old, p.Options.New, p.Options.Count)
}
