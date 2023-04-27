package process

import (
	"bytes"
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var (
	_ Applier  = procMath{}
	_ Batcher  = procMath{}
	_ Streamer = procMath{}
)

var mathTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected []byte
	err      error
}{
	{
		"add",
		config.Config{
			Type: "add",
			Settings: map[string]interface{}{
				"key":     "math",
				"set_key": "math",
				"options": map[string]interface{}{
					"operation": "add",
				},
			},
		},
		[]byte(`{"math":[1,3]}`),
		[]byte(`{"math":4}`),
		nil,
	},
	{
		"subtract",
		config.Config{
			Type: "add",
			Settings: map[string]interface{}{
				"key":     "math",
				"set_key": "math",
				"options": map[string]interface{}{
					"operation": "subtract",
				},
			},
		},
		[]byte(`{"math":[5,2]}`),
		[]byte(`{"math":3}`),
		nil,
	},
	{
		"multiply",
		config.Config{
			Type: "add",
			Settings: map[string]interface{}{
				"key":     "math",
				"set_key": "math",
				"options": map[string]interface{}{
					"operation": "multiply",
				},
			},
		},
		[]byte(`{"math":[10,2]}`),
		[]byte(`{"math":20}`),
		nil,
	},
	{
		"divide",
		config.Config{
			Type: "add",
			Settings: map[string]interface{}{
				"key":     "math",
				"set_key": "math",
				"options": map[string]interface{}{
					"operation": "divide",
				},
			},
		},
		[]byte(`{"math":[10,2]}`),
		[]byte(`{"math":5}`),
		nil,
	},
}

func TestMath(t *testing.T) {
	ctx := context.TODO()
	capsule := config.NewCapsule()

	for _, test := range mathTests {
		t.Run(test.name, func(t *testing.T) {
			capsule.SetData(test.test)

			proc, err := newProcMath(ctx, test.cfg)
			if err != nil {
				t.Fatal(err)
			}

			result, err := proc.Apply(ctx, capsule)
			if err != nil {
				t.Error(err)
			}

			if !bytes.Equal(result.Data(), test.expected) {
				t.Errorf("expected %s, got %s", test.expected, result.Data())
			}
		})
	}
}

func benchmarkMath(b *testing.B, applier procMath, test config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = applier.Apply(ctx, test)
	}
}

func BenchmarkMath(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range mathTests {
		proc, err := newProcMath(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkMath(b, proc, capsule)
			},
		)
	}
}
