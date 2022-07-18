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
	Old   string `json:"old"`
	New   string `json:"new"`
	Count int    `json:"count"`
}

/*
Replace processes data by replacing characters. The processor supports these patterns:
	JSON:
		{"replace":"bar"} >>> {"replace":"baz"}
	data:
		bar >>> baz

The processor uses this Jsonnet configuration:
	{
		type: 'replace',
		settings: {
			options: {
				old: 'r',
				new: 'z',
			},
			input_key: 'replace',
			output_key: 'replace',
		},
	}
*/
type Replace struct {
	Options   ReplaceOptions           `json:"options"`
	Condition condition.OperatorConfig `json:"condition"`
	InputKey  string                   `json:"input_key"`
	OutputKey string                   `json:"output_key"`
}

// Slice processes a slice of bytes with the Replace processor. Conditions are optionally applied on the bytes to enable processing.
func (p Replace) Slice(ctx context.Context, s [][]byte) ([][]byte, error) {
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

// Byte processes bytes with the Replace processor.
func (p Replace) Byte(ctx context.Context, data []byte) ([]byte, error) {
	// error early if required options are missing
	if p.Options.Old == "" || p.Options.New == "" {
		return nil, fmt.Errorf("byter settings %+v: %w", p, ProcessorInvalidSettings)
	}

	// default to replace all
	if p.Options.Count == 0 {
		p.Options.Count = -1
	}

	// JSON processing
	if p.InputKey != "" && p.OutputKey != "" {
		value := json.Get(data, p.InputKey).String()
		rep := strings.Replace(value, p.Options.Old, p.Options.New, p.Options.Count)
		return json.Set(data, p.OutputKey, rep)
	}

	// data processing
	if p.InputKey == "" && p.OutputKey == "" {
		rep := bytes.Replace(data, []byte(p.Options.Old), []byte(p.Options.New), p.Options.Count)
		return rep, nil
	}

	return nil, fmt.Errorf("byter settings %+v: %w", p, ProcessorInvalidSettings)
}
