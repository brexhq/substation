package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
)

// insert processes data by inserting a value into an object.
//
// This processor supports the object handling pattern.
type procInsert struct {
	process
	Options procInsertOptions `json:"options"`
}

type procInsertOptions struct {
	// Value inserted into the object.
	Value interface{} `json:"value"`
}

// Create a new insert processor.
func newProcInsert(ctx context.Context, cfg config.Config) (p procInsert, err error) {
	if err = config.Decode(cfg.Settings, &p); err != nil {
		return procInsert{}, err
	}

	p.operator, err = condition.NewOperator(ctx, p.Condition)
	if err != nil {
		return procInsert{}, err
	}

	// only supports JSON, fail if there are no keys
	if p.SetKey == "" {
		return procInsert{}, fmt.Errorf("process: insert: set_key %s: %v", p.SetKey, errInvalidDataPattern)
	}

	return p, nil
}

// String returns the processor settings as an object.
func (p procInsert) String() string {
	return toString(p)
}

// Closes resources opened by the processor.
func (p procInsert) Close(context.Context) error {
	return nil
}

// Batch processes one or more capsules with the processor. Conditions are
// optionally applied to the data to enable processing.
func (p procInsert) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	return batchApply(ctx, capsules, p, p.operator)
}

// Apply processes a capsule with the processor.
func (p procInsert) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	if err := capsule.Set(p.SetKey, p.Options.Value); err != nil {
		return capsule, fmt.Errorf("process: insert: %v", err)
	}

	return capsule, nil
}
