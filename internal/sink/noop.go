package sink

import (
	"context"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/log"
)

// noop is a no-op sink that discards data.
type sinkNoop struct{}

// Create a new noop sink.
func newSinkNoop(_ context.Context, cfg config.Config) (s sinkNoop, err error) {
	return s, nil
}

// Send sinks a channel of encapsulated data with the sink.
func (s sinkNoop) Send(ctx context.Context, ch *config.Channel) error {
	var count int
	for range ch.C {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			count++
		}
	}

	log.WithField(
		"count", count,
	).Debug("discarded data")

	return nil
}
