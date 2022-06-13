package process

import (
	"bytes"
	"context"
	"fmt"
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
	Case:
		the case to convert the string or byte to
		must be one of:
			upper
			lower
			snake (strings only)
*/
type CaseOptions struct {
	Case string `json:"case"`
}

/*
Case processes data by changing the case of a string or byte slice. The processor supports these patterns:
	json:
		{"case":"foo"} >>> {"case":"FOO"}
	json array:
		{"case":["foo","bar"]} >>> {"case":["FOO","BAR"]}
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
	Condition condition.OperatorConfig `json:"condition"`
	Input     Input                    `json:"input"`
	Output    Output                   `json:"output"`
	Options   CaseOptions              `json:"options"`
}

// Slice processes a slice of bytes with the Case processor. Conditions are optionally applied on the bytes to enable processing.
func (p Case) Slice(ctx context.Context, s [][]byte) ([][]byte, error) {
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

// Byte processes bytes with the Case processor.
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

	// data processing
	if p.Input.Key == "" && p.Output.Key == "" {
		return p.bytesCase(data), nil
	}

	return nil, fmt.Errorf("byter settings %v: %v", p, CaseInvalidSettings)
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
