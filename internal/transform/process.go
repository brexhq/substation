package transform

import (
	"context"

	"github.com/brexhq/substation/process"
)

// Process contains this transform's configuration settings. More information is available in the README.
type Process struct {
	Processors []process.Config `mapstructure:"processors"`
}

// Transform processes a channel of bytes with processors defined in the config.
func (transform *Process) Transform(ctx context.Context, in <-chan []byte, out chan<- []byte, kill chan struct{}) error {
	channelers, err := process.MakeAllChannelers(transform.Processors)
	if err != nil {
		return err
	}

	ch := in
	ch, err = process.Channel(ctx, channelers, ch)
	if err != nil {
		return err
	}

	for data := range ch {
		select {
		case <-kill:
			return nil
		default:
			out <- data
		}
	}

	return nil
}
