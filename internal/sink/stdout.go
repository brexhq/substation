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
func (sink *Stdout) Send(ctx context.Context, ch chan config.Capsule, kill chan struct{}) error {
	for cap := range ch {
		select {
		case <-kill:
			return nil
		default:
			fmt.Println(string(cap.GetData()))
		}
	}

	return nil
}
