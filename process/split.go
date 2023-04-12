package process

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
)

// split processes data by splitting it into multiple elements in an object array, objects, or strings.
//
// This processor supports the data and object handling patterns.
type procSplit struct {
	process
	Options procSplitOptions `json:"options"`
}

type procSplitOptions struct {
	// Separator is the string that splits data.
	Separator string `json:"separator"`
}

// Create a new split processor.
func newProcSplit(ctx context.Context, cfg config.Config) (p procSplit, err error) {
	if err = config.Decode(cfg.Settings, &p); err != nil {
		return procSplit{}, err
	}

	p.operator, err = condition.NewOperator(ctx, p.Condition)
	if err != nil {
		return procSplit{}, err
	}

	// fail if required options are missing
	if p.Options.Separator == "" {
		return procSplit{}, fmt.Errorf("process: split: %v separator", errors.ErrMissingRequiredOption)
	}

	return p, nil
}

// String returns the processor settings as an object.
func (p procSplit) String() string {
	return toString(p)
}

// Closes resources opened by the processor.
func (p procSplit) Close(context.Context) error {
	return nil
}

// Batch processes one or more capsules with the processor. Conditions are
// optionally applied to the data to enable processing.
func (p procSplit) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	newCapsules := newBatch(&capsules)
	for _, capsule := range capsules {
		ok, err := p.operator.Operate(ctx, capsule)
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
func (p procSplit) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	result := capsule.Get(p.Key).String()
	value := strings.Split(result, p.Options.Separator)

	if err := capsule.Set(p.SetKey, value); err != nil {
		return capsule, fmt.Errorf("process: split: %v", err)
	}

	return capsule, nil
}
