package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
)

// copy processes data by copying it into, from, and inside objects.
//
// This processor supports the data and object handling patterns.
type procCopy struct {
	process
}

// Create a new copy processor.
func newProcCopy(ctx context.Context, cfg config.Config) (p procCopy, err error) {
	if err = config.Decode(cfg.Settings, &p); err != nil {
		return procCopy{}, err
	}

	p.operator, err = condition.NewOperator(ctx, p.Condition)
	if err != nil {
		return procCopy{}, err
	}

	return p, nil
}

// String returns the processor settings as an object.
func (p procCopy) String() string {
	return toString(p)
}

// Closes resources opened by the processor.
func (p procCopy) Close(context.Context) error {
	return nil
}

// Stream processes a pipeline of capsules with the processor.
func (p procCopy) Stream(ctx context.Context, in, out *config.Channel) error {
	return streamApply(ctx, in, out, p)
}

// Batch processes one or more capsules with the processor.
func (p procCopy) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	return batchApply(ctx, capsules, p)
}

// Apply processes a capsule with the processor.
func (p procCopy) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	if ok, err := p.operator.Operate(ctx, capsule); err != nil {
		return capsule, fmt.Errorf("process: copy: %v", err)
	} else if !ok {
		return capsule, nil
	}

	// JSON processing
	if p.Key != "" && p.SetKey != "" {
		if err := capsule.Set(p.SetKey, capsule.Get(p.Key)); err != nil {
			return capsule, fmt.Errorf("process: copy: %v", err)
		}

		return capsule, nil
	}

	// from JSON processing
	if p.Key != "" && p.SetKey == "" {
		result := capsule.Get(p.Key).String()

		capsule.SetData([]byte(result))
		return capsule, nil
	}

	// to JSON processing
	if p.Key == "" && p.SetKey != "" {
		if err := capsule.Set(p.SetKey, capsule.Data()); err != nil {
			return capsule, fmt.Errorf("process: copy: %v", err)
		}

		return capsule, nil
	}

	return capsule, fmt.Errorf("process: copy: key %s set_key %s: %w", p.Key, p.SetKey, errInvalidDataPattern)
}
