package secrets

import (
	"context"
	"testing"

	"github.com/brexhq/substation/v2/config"
)

func TestCollect(t *testing.T) {
	t.Setenv("FOO", "bar")

	ctx := context.Background()

	cfg := config.Config{
		Type: "environment_variable",
		Settings: map[string]interface{}{
			"id":   "id",
			"name": "FOO",
		},
	}

	ret, err := New(ctx, cfg)
	if err != nil {
		// handle error
		panic(err)
	}

	if err := ret.Retrieve(ctx); err != nil {
		// handle error
		panic(err)
	}

	interp, err := Interpolate(context.TODO(), "/path/to/secret/${SECRET:id}")
	if err != nil {
		// handle error
		panic(err)
	}

	if interp != "/path/to/secret/bar" {
		t.Fatalf("unexpected interpolation: %s", interp)
	}
}
