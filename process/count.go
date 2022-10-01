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
func (p Count) ApplyBatch(ctx context.Context, caps []config.Capsule) ([]config.Capsule, error) {
	newCap := config.NewCapsule()
	if err := newCap.Set("count", len(caps)); err != nil {
		return caps, fmt.Errorf("count apply: : %v", err)
	}

	newCaps := make([]config.Capsule, 0, 1)
	newCaps = append(newCaps, newCap)
	return newCaps, nil
}
