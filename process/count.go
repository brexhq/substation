package process

import (
	"context"
	"fmt"

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

// Slice processes a slice of bytes with the Count processor. Conditions are optionally applied on the bytes to enable processing.
func (p Count) Slice(ctx context.Context, s [][]byte) ([][]byte, error) {
	processed, err := json.Set([]byte(""), "count", len(s))
	if err != nil {
		return nil, fmt.Errorf("slicer settings %v: %v", p, err)
	}

	slice := make([][]byte, 1, 1)
	slice[0] = processed

	return slice, nil
}
