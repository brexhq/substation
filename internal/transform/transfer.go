package transform

import (
	"context"
)

// Transfer contains this transform's configuration settings. More information is available in the README.
type Transfer struct{}

// Transform tranfers the input channel of bytes directly to the output channel with no modification.
func (transform *Transfer) Transform(ctx context.Context, in <-chan []byte, out chan<- []byte, kill chan struct{}) error {
	for data := range in {
		select {
		case <-kill:
			return nil
		default:
			out <- data
		}
	}

	return nil
}
