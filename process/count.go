package process

import (
	"context"

	"github.com/brexhq/substation/internal/json"
)

/*
Count processes data by counting it.

The processor uses this Jsonnet configuration:
	{
		type: 'count',
	}
*/
type Count struct{}

// Channel processes a channel of byte slices with the Count processor.
func (p Count) Channel(ctx context.Context, ch <-chan []byte) (<-chan []byte, error) {
	output := make(chan []byte, 1)
	defer close(output)

	var count int
	for {
		_, ok := <-ch
		if !ok {
			break
		}
		count++
	}

	processed, err := json.Set([]byte(""), "count", count)
	if err != nil {
		return nil, err
	}
	output <- processed

	return output, nil
}
