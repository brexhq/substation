package process

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
)

// split processes data by splitting it into multiple elements in an object array, objects, or strings.
//
// This processor supports the data and object handling patterns.
type _split struct {
	process
	Options _splitOptions `json:"options"`
}

type _splitOptions struct {
	// Separator is the string that splits data.
	Separator string `json:"separator"`
}

// String returns the processor settings as an object.
func (p _split) String() string {
	return toString(p)
}

// Close closes resources opened by the processor.
func (p _split) Close(context.Context) error {
	return nil
}

// Batch processes one or more capsules with the processor. Conditions are
// optionally applied to the data to enable processing.
func (p _split) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	op, err := condition.MakeOperator(p.Condition)
	if err != nil {
		return nil, fmt.Errorf("process: split: %v", err)
	}

	newCapsules := newBatch(&capsules)
	for _, capsule := range capsules {
		ok, err := op.Operate(ctx, capsule)
		if err != nil {
			return nil, fmt.Errorf("process: split: %v", err)
		}

		if !ok {
			newCapsules = append(newCapsules, capsule)
			continue
		}

		// JSON processing
		if p.Key != "" && p.SetKey != "" {
			pcap, err := p.Apply(ctx, capsule)
			if err != nil {
				return nil, fmt.Errorf("process: split: %v", err)
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

		return nil, fmt.Errorf("process: split: key %s set_key %s: %v", p.Key, p.SetKey, errInvalidDataPattern)
	}

	return newCapsules, nil
}

// Apply processes a capsule with the processor.
func (p _split) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	// error early if required options are missing
	if p.Options.Separator == "" {
		return capsule, fmt.Errorf("process: split: options %+v: %v", p.Options, errMissingRequiredOptions)
	}

	// only supports JSON, error early if there are no keys
	if p.Key == "" || p.SetKey == "" {
		return capsule, fmt.Errorf("process: split: key %s set_key %s: %v", p.Key, p.SetKey, errInvalidDataPattern)
	}

	result := capsule.Get(p.Key).String()
	value := strings.Split(result, p.Options.Separator)

	if err := capsule.Set(p.SetKey, value); err != nil {
		return capsule, fmt.Errorf("process: split: %v", err)
	}

	return capsule, nil
}
