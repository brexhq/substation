package process

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/iancoleman/strcase"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/internal/json"
)

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
	JSON:
		{"case":"foo"} >>> {"case":"FOO"}
	data:
		foo >>> FOO

The processor uses this Jsonnet configuration:
	{
		type: 'case',
		settings: {
			options: {
				case: 'upper',
			},
			input_key: 'case',
			output_key: 'case',
		},
	}
*/
type Case struct {
	Options   CaseOptions              `json:"options"`
	Condition condition.OperatorConfig `json:"condition"`
	InputKey  string                   `json:"input_key"`
	OutputKey string                   `json:"output_key"`
}

// Slice processes a slice of bytes with the Case processor. Conditions are optionally applied on the bytes to enable processing.
func (p Case) Slice(ctx context.Context, s [][]byte) ([][]byte, error) {
	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, fmt.Errorf("slicer settings %+v: %w", p, err)
	}

	slice := NewSlice(&s)
	for _, data := range s {
		ok, err := op.Operate(data)
		if err != nil {
			return nil, fmt.Errorf("slicer settings %+v: %w", p, err)
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
	// error early if required options are missing
	if p.Options.Case == "" {
		return nil, fmt.Errorf("byter settings %+v: %w", p, ProcessorInvalidSettings)
	}

	// JSON processing
	if p.InputKey != "" && p.OutputKey != "" {
		value := json.Get(data, p.InputKey).String()
		switch p.Options.Case {
		case "upper":
			return json.Set(data, p.OutputKey, strings.ToUpper(value))
		case "lower":
			return json.Set(data, p.OutputKey, strings.ToLower(value))
		case "snake":
			return json.Set(data, p.OutputKey, strcase.ToSnake(value))
		}
	}

	// data processing
	if p.InputKey == "" && p.OutputKey == "" {
		switch p.Options.Case {
		case "upper":
			return bytes.ToUpper(data), nil
		case "lower":
			return bytes.ToLower(data), nil
		}
	}

	return nil, fmt.Errorf("byter settings %+v: %w", p, ProcessorInvalidSettings)
}
