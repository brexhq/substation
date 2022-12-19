package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/config"
)

type count struct{}

// Close closes resources opened by the Count processor.
func (p count) Close(context.Context) error {
	return nil
}

// ApplyBatch processes a slice of encapsulated data with the Count processor. Conditions are optionally applied to the data to enable processing.
func (p count) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	newCapsule := config.NewCapsule()
	if err := newCapsule.Set("count", len(capsules)); err != nil {
		return capsules, fmt.Errorf("process count: : %v", err)
	}

	newCapsules := make([]config.Capsule, 0, 1)
	newCapsules = append(newCapsules, newCapsule)
	return newCapsules, nil
}
