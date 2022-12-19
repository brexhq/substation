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
split processes data by splitting it into multiple elements or items. The processor supports these patterns:

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
type split struct {
	process
	Options splitOptions `json:"options"`
}

/*
splitOptions contains custom options settings for the split processor:

	Separator:
		string that separates aggregated data
*/
type splitOptions struct {
	Separator string `json:"separator"`
}

// Close closes resources opened by the split processor.
func (p split) Close(context.Context) error {
	return nil
}

// ApplyBatch processes a slice of encapsulated data with the split processor. Conditions are optionally applied to the data to enable processing.
func (p split) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, fmt.Errorf("process split: %v", err)
	}

	newCapsules := newBatch(&capsules)
	for _, capsule := range capsules {
		ok, err := op.Operate(ctx, capsule)
		if err != nil {
			return nil, fmt.Errorf("process split: %v", err)
		}

		if !ok {
			newCapsules = append(newCapsules, capsule)
			continue
		}

		// JSON processing
		if p.Key != "" && p.SetKey != "" {
			pcap, err := p.Apply(ctx, capsule)
			if err != nil {
				return nil, fmt.Errorf("process split: %v", err)
			}
			newCapsules = append(newCapsules, pcap)

			continue
		}

		// data processing
		if p.Key == "" && p.SetKey == "" {
			newCapsule := config.NewCapsule()
			for _, x := range bytes.Split(capsule.Data(), []byte(p.Options.Separator)) {
				newCapsule.SetData(x)
				newCapsules = append(newCapsules, newCapsule)
			}

			continue
		}

		return nil, fmt.Errorf("process split: inputkey %s outputkey %s: %v", p.Key, p.SetKey, errInvalidDataPattern)
	}

	return newCapsules, nil
}

// Apply processes encapsulated data with the split processor.
func (p split) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	// error early if required options are missing
	if p.Options.Separator == "" {
		return capsule, fmt.Errorf("process split: options %+v: %v", p.Options, errMissingRequiredOptions)
	}

	// only supports JSON, error early if there are no keys
	if p.Key == "" || p.SetKey == "" {
		return capsule, fmt.Errorf("process split: inputkey %s outputkey %s: %v", p.Key, p.SetKey, errInvalidDataPattern)
	}

	result := capsule.Get(p.Key).String()
	value := strings.Split(result, p.Options.Separator)

	if err := capsule.Set(p.SetKey, value); err != nil {
		return capsule, fmt.Errorf("process split: %v", err)
	}

	return capsule, nil
}
