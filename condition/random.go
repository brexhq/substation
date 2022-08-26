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

/*
Random evaluates data based on a random choice. This inspector uses the standard library's rand package. This is best paired with the drop processor and can be used to integration test new configurations and deployments.

When loaded with a factory, the inspector uses this JSON configuration:
	{
		"type": "random"
	}
*/
type Random struct{}

// Inspect evaluates encapsulated data with the Random inspector.
func (c Random) Inspect(ctx context.Context, cap config.Capsule) (output bool, err error) {
	matched := rand.Intn(2) == 1
	return matched, nil
}
