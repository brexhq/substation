package process

import (
	"context"

	"github.com/brexhq/substation/internal/json"
)

// Count implements the Channeler interfaces and counts all data put into the channel. More information is available in the README.
type Count struct{}

// Channel processes a data channel of bytes with this processor.
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
