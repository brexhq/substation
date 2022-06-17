package sink

import (
	"context"
	"fmt"
)

/*
Stdout sinks data to stdout.

The sink uses this Jsonnet configuration:
	{
		type: 'stdout',
	}
*/
type Stdout struct{}

// Send sinks a channel of bytes with the Stdout sink.
func (sink *Stdout) Send(ctx context.Context, ch chan []byte, kill chan struct{}) error {
	for data := range ch {
		select {
		case <-kill:
			return nil
		default:
			fmt.Println(string(data))
		}
	}

	return nil
}
