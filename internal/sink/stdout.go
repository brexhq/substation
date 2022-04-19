package sink

import (
	"context"
	"fmt"
)

// Stdout implements the Sink interface and prints data to stdout. More information is available in the README.
type Stdout struct{}

// Send prints a channel of bytes to stdout.
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
