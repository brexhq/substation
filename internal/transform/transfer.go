package transform

import (
	"context"
	"sync"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/metrics"
)

// transfer transforms data without modification.
type _transfer struct{}

// Transform processes a channel of encapsulated data with the transform.
func (transform *_transfer) Transform(ctx context.Context, wg *sync.WaitGroup, in, out *config.Channel) error {
	var count int

	// read and write encapsulated data from input and to output channels
	for capsule := range in.C {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			out.Send(capsule)
			count++
		}
	}

	_ = metrics.Generate(ctx, metrics.Data{
		Name:  "CapsulesReceived",
		Value: count,
	})

	_ = metrics.Generate(ctx, metrics.Data{
		Name:  "CapsulesSent",
		Value: count,
	})

	return nil
}
