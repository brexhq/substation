package sink

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/config"
)

/*
Stdout sinks data to stdout.

When loaded with a factory, the sink uses this JSON configuration:
	{
		"type": "stdout"
	}
*/
type Stdout struct{}

// Send sinks a channel of encapsulated data with the Stdout sink.
func (sink *Stdout) Send(ctx context.Context, ch *config.Channel) error {
	var count int
	for cap := range ch.C {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			fmt.Println(string(cap.Data()))
			count++
		}
	}

	return nil
}
