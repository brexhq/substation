package process

import (
	"bytes"
	"context"
	"strings"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/json"
)

// ReplaceInvalidSettings is returned when the Replace processor is configured with invalid Input and Output settings.
const ReplaceInvalidSettings = errors.Error("ReplaceInvalidSettings")

/*
ReplaceOptions contains custom options for the Replace processor:
	Old:
		the character(s) to replace in the data
	New:
		the character(s) that replace Old
	Count (optional):
		the number of replacements to make
		defaults to -1, which replaces all matches
*/
type ReplaceOptions struct {
	Old   string `mapstructure:"old"`
	New   string `mapstructure:"new"`
	Count int    `mapstructure:"count"`
}

/*
Replace processes data by replacing characters. The processor supports these patterns:
	json:
		{"replace":"bar"} >>> {"replace":"baz"}
	json array:
		{"replace":["bar","bard"]} >>> {"replace":["baz","bazd"]}
	data:
		bar >>> baz

The processor uses this Jsonnet configuration:
	{
		type: 'replace',
		settings: {
			input: {
				key: 'replace',
			},
			output: {
				key: 'replace',
			}
			options: {
				old: 'r',
				new: 'z',
			}
		},
	}
*/
type Replace struct {
	Condition condition.OperatorConfig `mapstructure:"condition"`
	Input     Input                    `mapstructure:"input"`
	Output    Output                   `mapstructure:"output"`
	Options   ReplaceOptions           `mapstructure:"options"`
}

// Channel processes a data channel of byte slices with the Replace processor. Conditions are optionally applied on the channel data to enable processing.
func (p Replace) Channel(ctx context.Context, ch <-chan []byte) (<-chan []byte, error) {
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

// Byte processes a byte slice with the Replace processor.
func (p Replace) Byte(ctx context.Context, data []byte) ([]byte, error) {
	// default to replace all
	if p.Options.Count == 0 {
		p.Options.Count = -1
	}

	// json processing
	if p.Input.Key != "" && p.Output.Key != "" {
		value := json.Get(data, p.Input.Key)
		if !value.IsArray() {
			r := p.stringsReplace(value.String())
			return json.Set(data, p.Output.Key, r)
		}

		// json array processing
		var array []string
		for _, v := range value.Array() {
			r := p.stringsReplace(v.String())
			array = append(array, r)
		}

		return json.Set(data, p.Output.Key, array)
	}

	// data processing
	if p.Input.Key == "" && p.Output.Key == "" {
		return p.bytesReplace(data), nil
	}

	return nil, ReplaceInvalidSettings
}

func (p Replace) stringsReplace(s string) string {
	return strings.Replace(s, p.Options.Old, p.Options.New, p.Options.Count)
}

func (p Replace) bytesReplace(b []byte) []byte {
	return bytes.Replace(b, []byte(p.Options.Old), []byte(p.Options.New), p.Options.Count)
}
