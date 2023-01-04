package sink

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/config"
)

// stdout sinks data to standard output.
type sinkStdout struct{}

// Send sinks a channel of encapsulated data with the sink.
func (s *sinkStdout) Send(ctx context.Context, ch *config.Channel) error {
	var count int
	for capsule := range ch.C {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			fmt.Println(string(capsule.Data()))
			count++
		}
	}

	return nil
}
