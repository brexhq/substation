package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/config"
)

/*
Count processes data by counting it.

When loaded with a factory, the processor uses this JSON configuration:

	{
		"type": "count"
	}
*/
type Count struct{}

// ApplyBatch processes a slice of encapsulated data with the Count processor. Conditions are optionally applied to the data to enable processing.
func (p Count) ApplyBatch(ctx context.Context, capsules []config.Capsule) ([]config.Capsule, error) {
	newCapsule := config.NewCapsule()
	if err := newCapsule.Set("count", len(capsules)); err != nil {
		return capsules, fmt.Errorf("process count: : %v", err)
	}

	newCapsules := make([]config.Capsule, 0, 1)
	newCapsules = append(newCapsules, newCapsule)
	return newCapsules, nil
}
