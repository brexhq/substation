package process

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
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
Replace processes encapsulated data by replacing characters. The processor supports these patterns:
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

// ApplyBatch processes a slice of encapsulated data with the Replcae processor. Conditions are optionally applied to the data to enable processing.
func (p Replace) ApplyBatch(ctx context.Context, caps []config.Capsule) ([]config.Capsule, error) {
	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, fmt.Errorf("applybatch settings %+v: %w", p, err)
	}

	caps, err = conditionallyApplyBatch(ctx, caps, op, p)
	if err != nil {
		return nil, fmt.Errorf("applybatch settings %+v: %w", p, err)
	}

	return caps, nil
}

// Apply processes encapsulated data with the Replace processor.
func (p Replace) Apply(ctx context.Context, cap config.Capsule) (config.Capsule, error) {
	// error early if required options are missing
	if p.Options.Old == "" || p.Options.New == "" {
		return cap, fmt.Errorf("applicator settings %+v: %w", p, ProcessorInvalidSettings)
	}

	// default to replace all
	if p.Options.Count == 0 {
		p.Options.Count = -1
	}

	// JSON processing
	if p.InputKey != "" && p.OutputKey != "" {
		result := cap.Get(p.InputKey).String()
		rep := strings.Replace(
			result,
			p.Options.Old,
			p.Options.New,
			p.Options.Count,
		)
		cap.Set(p.OutputKey, rep)

		return cap, nil
	}

	// data processing
	if p.InputKey == "" && p.OutputKey == "" {
		rep := bytes.Replace(
			cap.GetData(),
			[]byte(p.Options.Old),
			[]byte(p.Options.New),
			p.Options.Count,
		)
		cap.SetData(rep)

		return cap, nil
	}

	return cap, fmt.Errorf("applicator settings %+v: %w", p, ProcessorInvalidSettings)
}
