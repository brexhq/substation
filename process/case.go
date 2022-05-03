package process

import (
	"bytes"
	"context"
	"strings"

	"github.com/iancoleman/strcase"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/json"
)

// CaseInvalidSettings is returned when the Case processor is configured with invalid Input and Output settings.
const CaseInvalidSettings = errors.Error("CaseInvalidSettings")

/*
CaseOptions contains custom options for the Case processor:
	case:
		the case to convert the string or byte to
		must be one of:
			upper
			lower
			snake (strings only)
*/
type CaseOptions struct {
	Case string `mapstructure:"case"`
}

/*
Case processes data by changing the case of a string or byte slice. The processor supports these patterns:
	json:
		{"case":"foo"} >>> {"case":"FOO"}
	json array:
		{"capture":["foo","bar"]} >>> {"case":["FOO","BAR"]}
	from json:
		{"capture":"foo"} >>> FOO
	to json:
		foo >>> {"case":"FOO"}
	data:
		foo >>> FOO

The processor uses this Jsonnet configuration:
	{
		type: 'case',
		settings: {
			// if the value is "foo", then this returns "FOO"
			input: {
				key: 'case',
			},
			output: {
				key: 'case',
			},
			options: {
				case: 'upper',
			}
		},
	}
*/
type Case struct {
	Condition condition.OperatorConfig `mapstructure:"condition"`
	Input     Input                    `mapstructure:"input"`
	Output    Output                   `mapstructure:"output"`
	Options   CaseOptions              `mapstructure:"options"`
}

// Channel processes a data channel of bytes with this processor. Conditions can be optionally applied on the channel data to enable processing.
func (p Case) Channel(ctx context.Context, ch <-chan []byte) (<-chan []byte, error) {
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
func (p Case) Byte(ctx context.Context, data []byte) ([]byte, error) {
	// json processing
	if p.Input.Key != "" && p.Output.Key != "" {
		value := json.Get(data, p.Input.Key)
		if !value.IsArray() {
			s := p.stringsCase(value.String())
			return json.Set(data, p.Output.Key, s)
		}
		// json array processing
		var array []string
		for _, v := range value.Array() {
			s := p.stringsCase(v.String())
			array = append(array, s)
		}

		return json.Set(data, p.Output.Key, array)
	}

	// from json processing
	if p.Input.Key != "" && p.Output.Key == "" {
		v := json.Get(data, p.Input.Key)
		b := []byte(v.String())
		return p.bytesCase(b), nil
	}

	// to json processing
	if p.Input.Key == "" && p.Output.Key != "" {
		return json.Set(data, p.Output.Key, p.bytesCase(data))
	}

	// data processing
	if p.Input.Key == "" && p.Output.Key == "" {
		return p.bytesCase(data), nil
	}

	return nil, CaseInvalidSettings
}

func (p Case) stringsCase(s string) string {
	switch t := p.Options.Case; t {
	case "upper":
		return strings.ToUpper(s)
	case "lower":
		return strings.ToLower(s)
	case "snake":
		return strcase.ToSnake(s)
	default:
		return ""
	}
}

func (p Case) bytesCase(b []byte) []byte {
	switch t := p.Options.Case; t {
	case "upper":
		return bytes.ToUpper(b)
	case "lower":
		return bytes.ToLower(b)
	default:
		return nil
	}
}
