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
Replace processes data by replacing characters. The processor supports these patterns:

	JSON:
		{"replace":"bar"} >>> {"replace":"baz"}
	data:
		bar >>> baz

When loaded with a factory, the processor uses this JSON configuration:

	{
		"type": "replace",
		"settings": {
			"options": {
				"old": "r",
				"new": "z"
			},
			"input_key": "replace",
			"output_key": "replace"
		}
	}
*/
type Replace struct {
	Options   ReplaceOptions   `json:"options"`
	Condition condition.Config `json:"condition"`
	InputKey  string           `json:"input_key"`
	OutputKey string           `json:"output_key"`
}

/*
ReplaceOptions contains custom options for the Replace processor:

	Old:
		character(s) to replace in the data
	New:
		character(s) that replace Old
	Count (optional):
		number of replacements to make
		defaults to -1, which replaces all matches
*/
type ReplaceOptions struct {
	Old   string `json:"old"`
	New   string `json:"new"`
	Count int    `json:"count"`
}

// ApplyBatch processes a slice of encapsulated data with the Replace processor. Conditions are optionally applied to the data to enable processing.
func (p Replace) ApplyBatch(ctx context.Context, capsules []config.Capsule) ([]config.Capsule, error) {
	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, fmt.Errorf("process replace: %v", err)
	}

	capsules, err = conditionallyApplyBatch(ctx, capsules, op, p)
	if err != nil {
		return nil, fmt.Errorf("process replace: %v", err)
	}

	return capsules, nil
}

// Apply processes encapsulated data with the Replace processor.
func (p Replace) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	// error early if required options are missing
	if p.Options.Old == "" {
		return capsule, fmt.Errorf("process replace: options %+v: %w", p.Options, errMissingRequiredOptions)
	}

	// default to replace all
	if p.Options.Count == 0 {
		p.Options.Count = -1
	}

	// JSON processing
	if p.InputKey != "" && p.OutputKey != "" {
		result := capsule.Get(p.InputKey).String()
		value := strings.Replace(
			result,
			p.Options.Old,
			p.Options.New,
			p.Options.Count,
		)

		if err := capsule.Set(p.OutputKey, value); err != nil {
			return capsule, fmt.Errorf("process replace: %v", err)
		}

		return capsule, nil
	}

	// data processing
	if p.InputKey == "" && p.OutputKey == "" {
		value := bytes.Replace(
			capsule.Data(),
			[]byte(p.Options.Old),
			[]byte(p.Options.New),
			p.Options.Count,
		)
		capsule.SetData(value)

		return capsule, nil
	}

	return capsule, fmt.Errorf("process replace: inputkey %s outputkey %s: %v", p.InputKey, p.OutputKey, errInvalidDataPattern)
}
