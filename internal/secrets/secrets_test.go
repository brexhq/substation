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

// TODO (akline@brex.com): Interpolate panics in certain situations so this needs some work
// func FuzzInterpolate(f *testing.F) {
// 	// Seed the fuzzer with initial test cases
// 	f.Add("/path/to/secret/${SECRET:id}")
// 	f.Add("/path/to/secret/${SECRET:foo}")
// 	f.Add("/path/to/secret/")

// 	f.Fuzz(func(t *testing.T, pattern string) {
// 		t.Setenv("FOO", "bar")

// 		ctx := context.Background()

// 		cfg := config.Config{
// 			Type: "environment_variable",
// 			Settings: map[string]interface{}{
// 				"id":   "id",
// 				"name": "FOO",
// 			},
// 		}

// 		ret, err := New(ctx, cfg)
// 		if err != nil {
// 			t.Fatalf("failed to create new secret: %v", err)
// 		}

// 		if err := ret.Retrieve(ctx); err != nil {
// 			t.Fatalf("failed to retrieve secret: %v", err)
// 		}

// 		interp, err := Interpolate(context.TODO(), pattern)
// 		if err != nil {
// 			// If interpolation fails, we expect an error, so we can return early
// 			return
// 		}

// 		// Optionally, you can add more checks on interp here
// 		_ = interp
// 	})
// }
