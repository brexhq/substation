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
func (transform *Transfer) Transform(ctx context.Context, in *config.Channel, out *config.Channel) error {
	var count int

	// read and write encapsulated data from input and to output channels
	// if a signal is received on the kill channel, then this is interrupted
	for cap := range in.C {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			out.Send(cap)
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
