package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
)

// delete processes data by deleting keys from an object.
//
// This processor supports the object handling pattern.
type procDelete struct {
	process
}

// Create a new delete processor.
func newProcDelete(ctx context.Context, cfg config.Config) (p procDelete, err error) {
	if err = config.Decode(cfg.Settings, &p); err != nil {
		return procDelete{}, err
	}

	p.operator, err = condition.NewOperator(ctx, p.Condition)
	if err != nil {
		return procDelete{}, err
	}

	if p.Key == "" {
		return procDelete{}, fmt.Errorf("process: delete: key %q: %v", p.Key, errInvalidDataPattern)
	}
	return p, nil
}

// String returns the processor settings as an object.
func (p procDelete) String() string {
	return toString(p)
}

// Closes resources opened by the processor.
func (p procDelete) Close(context.Context) error {
	return nil
}

// Stream processes a pipeline of capsules with the processor.
func (p procDelete) Stream(ctx context.Context, in, out *config.Channel) error {
	return streamApply(ctx, in, out, p)
}

// Batch processes one or more capsules with the processor.
func (p procDelete) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	return batchApply(ctx, capsules, p)
}

// Apply processes a capsule with the processor.
func (p procDelete) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	if ok, err := p.operator.Operate(ctx, capsule); err != nil {
		return capsule, fmt.Errorf("process: delete: %v", err)
	} else if !ok {
		return capsule, nil
	}

	if err := capsule.Delete(p.Key); err != nil {
		return capsule, fmt.Errorf("process: delete: %v", err)
	}

	return capsule, nil
}
