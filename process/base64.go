package process

import (
	"context"
	"unicode/utf8"

	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/internal/base64"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/json"
)

// Base64JSONDecodedBinary is returned when the Base64 processor is configured to decode output to JSON, but the output contains binary data and cannot be written as valid JSON.
const Base64JSONDecodedBinary = errors.Error("Base64JSONDecodedBinary")

/*
Base64Options contains custom options for the Base64 processor:
	Direction:
		the direction of the encoding
		must be one of:
			to: encode to base64
			from: decode from base64
*/
type Base64Options struct {
	Direction string `json:"direction"`
}

/*
Base64 processes data by converting it to and from base64. The processor supports these patterns:
	JSON:
	  	{"base64":"Zm9v"} >>> {"base64":"foo"}
	data:
		Zm9v >>> foo

The processor uses this Jsonnet configuration:
	{
		type: 'base64',
		settings: {
			input_key: 'base64',
			output_key: 'base64',
			options: {
				direction: 'from',
			}
		},
	}
*/
type Base64 struct {
	Options   Base64Options            `json:"options"`
	Condition condition.OperatorConfig `json:"condition"`
	InputKey  string                   `json:"input_key"`
	OutputKey string                   `json:"output_key"`
}

// Slice processes a slice of bytes with the Base64 processor. Conditions are optionally applied on the bytes to enable processing.
func (p Base64) Slice(ctx context.Context, s [][]byte) ([][]byte, error) {
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

// Byte processes bytes with the Base64 processor.
func (p Base64) Byte(ctx context.Context, data []byte) ([]byte, error) {
	// error early if required options are missing
	if p.Options.Direction == "" {
		return nil, fmt.Errorf("byter settings %+v: %v", p, ProcessorInvalidSettings)
	}

	// JSON processing
	if p.InputKey != "" && p.OutputKey != "" {
		value := json.Get(data, p.InputKey)
		tmp := []byte(value.String())

		switch p.Options.Direction {
		case "from":
			result, err := base64.Decode(tmp)
			if err != nil {
				return nil, fmt.Errorf("byter settings %v: %v", p, err)
			}

			if !utf8.Valid(result) {
				return nil, fmt.Errorf("byter settings %v: %v", p, Base64JSONDecodedBinary)
			}

			return json.Set(data, p.OutputKey, result)
		case "to":
			result := base64.Encode(tmp)
			return json.Set(data, p.OutputKey, string(result))
		default:
			return nil, fmt.Errorf("byter settings %v: %v", p, ProcessorInvalidDirection)
		}
	}

	// data processing
	if p.InputKey == "" && p.OutputKey == "" {
		switch p.Options.Direction {
		case "from":
			result, err := base64.Decode(data)
			if err != nil {
				return nil, fmt.Errorf("byter settings %v: %v", p, err)
			}
			return result, nil
		case "to":
			return base64.Encode(data), nil
		default:
			return nil, fmt.Errorf("byter settings %v: %v", p, ProcessorInvalidDirection)
		}
	}

	return nil, fmt.Errorf("byter settings %v: %v", p, ProcessorInvalidSettings)
}
