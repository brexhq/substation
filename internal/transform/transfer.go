package transform

import (
	"context"

	"github.com/brexhq/substation/config"
)

/*
Transfer transforms data without modification. This transform should be used when data needs to be moved from a source to a sink without processing data.

The transform uses this Jsonnet configuration:
	{
		type: 'transfer',
	}
*/
type Transfer struct{}

// Transform processes a channel of encapsulated data with the Transfer transform.
func (transform *Transfer) Transform(ctx context.Context, in <-chan config.Capsule, out chan<- config.Capsule, kill chan struct{}) error {
	// read and write encapsulated data from and to channels
	for cap := range in {
		select {
		case <-kill:
			return nil
		default:
			out <- cap
		}
	}

	return nil
}
