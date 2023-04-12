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

// replace processes data by replacing characters in strings.
//
// This processor supports the data and object handling patterns.
type procReplace struct {
	process
	Options procReplaceOptions `json:"options"`
}

type procReplaceOptions struct {
	// Old contains characters to replace in the data.
	Old string `json:"old"`
	// New contains characters that replace characters in Old.
	New string `json:"new"`
	// Counter determines the number of replacements to make.
	//
	// This is optional and defaults to -1 (replaces all matches).
	Count int `json:"count"`
}

// Create a new replace processor.
func newProcReplace(cfg config.Config) (p procReplace, err error) {
	if err = config.Decode(cfg.Settings, &p); err != nil {
		return procReplace{}, err
	}

	p.operator, err = condition.NewOperator(p.Condition)
	if err != nil {
		return procReplace{}, err
	}

	if p.Options.Old == "" {
		return procReplace{}, fmt.Errorf("process: replace: %v old", errors.ErrMissingRequiredOption)
	}

	// default to procReplace all
	if p.Options.Count == 0 {
		p.Options.Count = -1
	}

	return p, nil
}

// String returns the processor settings as an object.
func (p procReplace) String() string {
	return toString(p)
}

// Closes resources opened by the processor.
func (p procReplace) Close(context.Context) error {
	return nil
}

// Batch processes one or more capsules with the processor. Conditions are
// optionally applied to the data to enable processing.
func (p procReplace) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	return batchApply(ctx, capsules, p, p.operator)
}

// Apply processes a capsule with the processor.
func (p procReplace) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	// JSON processing
	if p.Key != "" && p.SetKey != "" {
		result := capsule.Get(p.Key).String()
		value := strings.Replace(
			result,
			p.Options.Old,
			p.Options.New,
			p.Options.Count,
		)

		if err := capsule.Set(p.SetKey, value); err != nil {
			return capsule, fmt.Errorf("process: replace: %v", err)
		}

		return capsule, nil
	}

	// data processing
	if p.Key == "" && p.SetKey == "" {
		value := bytes.Replace(
			capsule.Data(),
			[]byte(p.Options.Old),
			[]byte(p.Options.New),
			p.Options.Count,
		)
		capsule.SetData(value)

		return capsule, nil
	}

	return capsule, fmt.Errorf("process: replace: key %s set_key %s: %v", p.Key, p.SetKey, errInvalidDataPattern)
}
