package transform

import (
	"context"
)

/*
Transfer transforms data without modification. This transform should be used when data needs to be moved from a source to a sink without data processing.

The transform uses this Jsonnet configuration:
	{
		type: 'transfer',
	}
*/
type Transfer struct{}

// Transform processes a channel of bytes with the Transfer transform.
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
