package process

import (
	"context"

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
	newCap.Set("count", len(caps))

	newCaps := make([]config.Capsule, 1, 1)
	newCaps = append(newCaps, newCap)
	return newCaps, nil
}
