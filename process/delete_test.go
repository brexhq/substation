package process

import (
	"bytes"
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var (
	_ Applier = procDelete{}
	_ Batcher = procDelete{}
)

var deleteTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected []byte
	err      error
}{
	{
		"string",
		config.Config{
			Type: "delete",
			Settings: map[string]interface{}{
				"key": "baz",
			},
		},
		[]byte(`{"foo":"bar","baz":"qux"}`),
		[]byte(`{"foo":"bar"}`),
		nil,
	},
	{
		"JSON",
		config.Config{
			Type: "delete",
			Settings: map[string]interface{}{
				"key": "baz",
			},
		},
		[]byte(`{"foo":"bar","baz":{"qux":"quux"}}`),
		[]byte(`{"foo":"bar"}`),
		nil,
	},
}

func TestDelete(t *testing.T) {
	ctx := context.TODO()
	capsule := config.NewCapsule()

	for _, test := range deleteTests {
		t.Run(test.name, func(t *testing.T) {
			capsule.SetData(test.test)

			proc, err := newProcDelete(ctx, test.cfg)
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

func benchmarkDelete(b *testing.B, applier procDelete, test config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = applier.Apply(ctx, test)
	}
}

func BenchmarkDelete(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range deleteTests {
		proc, err := newProcDelete(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkDelete(b, proc, capsule)
			},
		)
	}
}
