package sink

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
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
func (sink *Stdout) Send(ctx context.Context, ch chan config.Capsule, kill chan struct{}) error {
	var count int
	for cap := range ch {
		select {
		case <-ctx.Done():
			return nil
		default:

			// used with errgroup branch, simulates failure
			if count == 10 {
				fmt.Println("stdout error")
				return errors.Error("stdout")
			}

			cap.GetData()
			// fmt.Println(string(cap.GetData()))

			count++
		}
	}

	return nil
}
