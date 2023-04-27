package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
)

// drop processes data by removing and not emitting it.
//
// This processor supports the data and object handling patterns.
type procDrop struct {
	process
}

// Create a new drop processor.
func newProcDrop(ctx context.Context, cfg config.Config) (p procDrop, err error) {
	if err = config.Decode(cfg.Settings, &p); err != nil {
		return procDrop{}, err
	}

	p.operator, err = condition.NewOperator(ctx, p.Condition)
	if err != nil {
		return procDrop{}, err
	}

	return p, nil
}

// String returns the processor settings as an object.
func (p procDrop) String() string {
	return toString(p)
}

// Closes resources opened by the processor.
func (p procDrop) Close(context.Context) error {
	return nil
}

// Stream processes a pipeline of capsules with the processor.
func (p procDrop) Stream(ctx context.Context, in, out *config.Channel) error {
	defer out.Close()

	for capsule := range in.C {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if ok, err := p.operator.Operate(ctx, capsule); err != nil {
				return fmt.Errorf("process: drop: %v", err)
			} else if !ok {
				out.Send(capsule)
				continue
			}
		}
	}

	return nil
}

// Batch processes one or more capsules with the processor.
func (p procDrop) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	newCapsules := newBatch(&capsules)
	for _, capsule := range capsules {
		if ok, err := p.operator.Operate(ctx, capsule); err != nil {
			return nil, fmt.Errorf("process: drop: %v", err)
		} else if !ok {
			newCapsules = append(newCapsules, capsule)
			continue
		}
	}

	return newCapsules, nil
}
