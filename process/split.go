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
Split processes data by splitting it into multiple elements or items. The processor supports these patterns:
	JSON:
		{"split":"foo.bar"} >>> {"split":["foo","bar"]}
	data:
		foo\nbar\nbaz\qux >>> foo bar baz qux
		{"foo":"bar"}\n{"baz":"qux"} >>> {"foo":"bar"} {"baz":"qux"}

When loaded with a factory, the processor uses this JSON configuration:
	{
		"type": "split",
		"settings": {
			"options": {
				"separator": "."
			},
			"input_key": "split",
			"output_key": "split"
		}
	}
*/
type Split struct {
	Options   SplitOptions     `json:"options"`
	Condition condition.Config `json:"condition"`
	InputKey  string           `json:"input_key"`
	OutputKey string           `json:"output_key"`
}

/*
SplitOptions contains custom options settings for the Split processor:
	Separator:
		the string that separates aggregated data
*/
type SplitOptions struct {
	Separator string `json:"separator"`
}

// ApplyBatch processes a slice of encapsulated data with the Split processor. Conditions are optionally applied to the data to enable processing.
func (p Split) ApplyBatch(ctx context.Context, caps []config.Capsule) ([]config.Capsule, error) {
	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, fmt.Errorf("applybatch settings %+v: %w", p, err)
	}

	newCaps := newBatch(&caps)
	for _, cap := range caps {
		ok, err := op.Operate(cap)
		if err != nil {
			return nil, fmt.Errorf("applybatch settings %+v: %w", p, err)
		}

		if !ok {
			newCaps = append(newCaps, cap)
			continue
		}

		// JSON processing
		if p.InputKey != "" && p.OutputKey != "" {
			pcap, err := p.Apply(ctx, cap)
			if err != nil {
				return nil, fmt.Errorf("applybatch: %v", err)
			}
			newCaps = append(newCaps, pcap)

			continue
		}

		// data processing
		if p.InputKey == "" && p.OutputKey == "" {
			newCap := config.NewCapsule()
			for _, x := range bytes.Split(cap.GetData(), []byte(p.Options.Separator)) {
				newCap.SetData(x)
				newCaps = append(newCaps, newCap)
			}

			continue
		}

		return nil, fmt.Errorf("applybatch settings %+v: %w", p, ProcessorInvalidSettings)
	}

	return newCaps, nil
}

// Apply processes encapsulated data with the Split processor.
func (p Split) Apply(ctx context.Context, cap config.Capsule) (config.Capsule, error) {
	// error early if required options are missing
	if p.Options.Separator == "" {
		return cap, fmt.Errorf("byter settings %+v: %w", p, ProcessorInvalidSettings)
	}

	// only supports JSON, error early if there are no keys
	if p.InputKey == "" || p.OutputKey == "" {
		return cap, fmt.Errorf("byter settings %+v: %w", p, ProcessorInvalidSettings)
	}

	str := cap.Get(p.InputKey).String()
	split := strings.Split(str, p.Options.Separator)

	cap.Set(p.OutputKey, split)
	return cap, nil
}
