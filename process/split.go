package process

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/internal/json"
)

/*
SplitOptions contains custom options settings for the Split processor:
	Separator:
		the string that separates aggregated data
*/
type SplitOptions struct {
	Separator string `json:"separator"`
}

/*
Split processes data by splitting it into multiple elements or items. The processor supports these patterns:
	JSON:
		{"split":"foo.bar"} >>> {"split":["foo","bar"]}
	data:
		foo\nbar\nbaz\qux >>> foo bar baz qux
		{"foo":"bar"}\n{"baz":"qux"} >>> {"foo":"bar"} {"baz":"qux"}

The processor uses this Jsonnet configuration:
	{
		type: 'split',
		settings: {
			options: {
				separator: '.',
			},
			input_key: 'split',
			output_key: 'split',
		},
	}
*/
type Split struct {
	Options   SplitOptions             `json:"options"`
	Condition condition.OperatorConfig `json:"condition"`
	InputKey  string                   `json:"input_key"`
	OutputKey string                   `json:"output_key"`
}

// Slice processes a slice of bytes with the Split processor. Conditions are optionally applied on the bytes to enable processing.
func (p Split) Slice(ctx context.Context, s [][]byte) ([][]byte, error) {
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

		// JSON processing
		if p.InputKey != "" && p.OutputKey != "" {
			processed, err := p.Byte(ctx, data)
			if err != nil {
				return nil, fmt.Errorf("slicer: %v", err)
			}
			slice = append(slice, processed)

			continue
		}

		// data processing
		if p.InputKey == "" && p.OutputKey == "" {
			for _, x := range bytes.Split(data, []byte(p.Options.Separator)) {
				slice = append(slice, x)
			}

			continue
		}

		return nil, fmt.Errorf("slicer settings %+v: %w", p, ProcessorInvalidSettings)
	}

	return slice, nil
}

// Byte processes bytes with the Split processor.
func (p Split) Byte(ctx context.Context, data []byte) ([]byte, error) {
	// error early if required options are missing
	if p.Options.Separator == "" {
		return nil, fmt.Errorf("byter settings %+v: %w", p, ProcessorInvalidSettings)
	}

	// only supports JSON, error early if there are no keys
	if p.InputKey == "" || p.OutputKey == "" {
		return nil, fmt.Errorf("byter settings %+v: %w", p, ProcessorInvalidSettings)
	}

	value := json.Get(data, p.InputKey).String()
	split := strings.Split(value, p.Options.Separator)
	return json.Set(data, p.OutputKey, split)
}
