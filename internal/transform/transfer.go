package transform

import (
	"context"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/metrics"
)

/*
Transfer transforms data without modification. This transform should be used when data needs to be moved from a source to a sink without processing.

When loaded with a factory, the transform uses this JSON configuration:
	{
		"type": "transfer"
	}
*/
type Transfer struct{}

// Transform processes a channel of encapsulated data with the Transfer transform.
func (transform *Transfer) Transform(ctx context.Context, in <-chan config.Capsule, out chan<- config.Capsule, kill chan struct{}) error {
	var count int

	// read and write encapsulated data from and to channels
	for cap := range in {
		select {
		case <-kill:
			return nil
		default:
			out <- cap
			count++
		}
	}

	metrics.Generate(ctx, metrics.Data{
		Name:  "CapsulesReceived",
		Value: count,
	})

	metrics.Generate(ctx, metrics.Data{
		Name:  "CapsulesSent",
		Value: count,
	})

	return nil
}
