package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/config"
)

// count processes data by counting it.
//
// This processor supports the data and object handling patterns.
type procCount struct{}

// Closes resources opened by the processor.
func (p procCount) Close(context.Context) error {
	return nil
}

// String returns the processor settings as an object.
func (p procCount) String() string {
	return toString(p)
}

// Batch processes one or more capsules with the processor. Conditions are
// optionally applied to the data to enable processing.
func (p procCount) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	newCapsule := config.NewCapsule()
	if err := newCapsule.Set("count", len(capsules)); err != nil {
		return capsules, fmt.Errorf("process: count: : %v", err)
	}

	newCapsules := make([]config.Capsule, 0, 1)
	newCapsules = append(newCapsules, newCapsule)
	return newCapsules, nil
}
