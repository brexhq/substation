package process

import (
	"bytes"
	"context"
	"fmt"
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
	Old   string `json:"old"`
	New   string `json:"new"`
	Count int    `json:"count"`
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
			input_key: 'replace',
			output_key: 'replace',
			options: {
				old: 'r',
				new: 'z',
			}
		},
	}
*/
type Replace struct {
	Condition condition.OperatorConfig `json:"condition"`
	InputKey  string                   `json:"input_key"`
	OutputKey string                   `json:"output_key"`
	Options   ReplaceOptions           `json:"options"`
}

// Slice processes a slice of bytes with the Replace processor. Conditions are optionally applied on the bytes to enable processing.
func (p Replace) Slice(ctx context.Context, s [][]byte) ([][]byte, error) {
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

// Byte processes bytes with the Replace processor.
func (p Replace) Byte(ctx context.Context, data []byte) ([]byte, error) {
	// default to replace all
	if p.Options.Count == 0 {
		p.Options.Count = -1
	}

	// json processing
	if p.InputKey != "" && p.OutputKey != "" {
		value := json.Get(data, p.InputKey)
		if !value.IsArray() {
			r := p.stringsReplace(value.String())
			return json.Set(data, p.OutputKey, r)
		}

		// json array processing
		var array []string
		for _, v := range value.Array() {
			r := p.stringsReplace(v.String())
			array = append(array, r)
		}

		return json.Set(data, p.OutputKey, array)
	}

	// data processing
	if p.InputKey == "" && p.OutputKey == "" {
		return p.bytesReplace(data), nil
	}

	return nil, fmt.Errorf("byter settings %v: %v", p, ReplaceInvalidSettings)
}

func (p Replace) stringsReplace(s string) string {
	return strings.Replace(s, p.Options.Old, p.Options.New, p.Options.Count)
}

func (p Replace) bytesReplace(b []byte) []byte {
	return bytes.Replace(b, []byte(p.Options.Old), []byte(p.Options.New), p.Options.Count)
}
