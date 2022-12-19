package condition

import (
	"context"
	"math/rand"
	"time"

	"github.com/brexhq/substation/config"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// random evaluates data based on a random choice using the standard library's rand package.
//
// This inspector supports the data and object handling patterns.
type random struct{}

// Inspect evaluates encapsulated data with the random inspector.
func (c random) Inspect(ctx context.Context, capsule config.Capsule) (output bool, err error) {
	return rand.Intn(2) == 1, nil
}
