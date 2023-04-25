package process

import (
	"bytes"
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var (
	_ Applier  = procFlatten{}
	_ Batcher  = procFlatten{}
	_ Streamer = procFlatten{}
)

var flattenTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected []byte
	err      error
}{
	{
		"json",
		config.Config{
			Type: "flatten",
			Settings: map[string]interface{}{
				"key":     "flatten",
				"set_key": "flatten",
			},
		},
		[]byte(`{"flatten":["foo",["bar"]]}`),
		[]byte(`{"flatten":["foo","bar"]}`),
		nil,
	},
	{
		"json deep flatten",
		config.Config{
			Type: "flatten",
			Settings: map[string]interface{}{
				"key":     "flatten",
				"set_key": "flatten",
				"options": map[string]interface{}{
					"deep": true,
				},
			},
		},
		[]byte(`{"flatten":[["foo"],[[["bar",[["baz"]]]]]]}`),
		[]byte(`{"flatten":["foo","bar","baz"]}`),
		nil,
	},
}

func TestFlatten(t *testing.T) {
	ctx := context.TODO()
	capsule := config.NewCapsule()

	for _, test := range flattenTests {
		t.Run(test.name, func(t *testing.T) {
			capsule.SetData(test.test)

			proc, err := newProcFlatten(ctx, test.cfg)
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

func benchmarkFlatten(b *testing.B, applier procFlatten, test config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = applier.Apply(ctx, test)
	}
}

func BenchmarkFlatten(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range flattenTests {
		proc, err := newProcFlatten(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkFlatten(b, proc, capsule)
			},
		)
	}
}
