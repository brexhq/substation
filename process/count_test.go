package process

import (
	"bytes"
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var _ Batcher = procCount{}

var countTests = []struct {
	name     string
	cfg      config.Config
	test     [][]byte
	expected []byte
	err      error
}{
	{
		"count",
		config.Config{},
		[][]byte{
			[]byte(`{"foo":"bar"}`),
			[]byte(`{"foo":"baz"}`),
			[]byte(`{"foo":"qux"}`),
		},
		[]byte(`{"count":3}`),
		nil,
	},
}

func TestCount(t *testing.T) {
	ctx := context.TODO()
	capsule := config.NewCapsule()

	for _, test := range countTests {
		t.Run(test.name, func(t *testing.T) {
			var capsules []config.Capsule
			for _, t := range test.test {
				capsule.SetData(t)
				capsules = append(capsules, capsule)
			}

			proc, err := newProcCount(test.cfg)
			if err != nil {
				t.Fatal(err)
			}

			result, err := proc.Batch(ctx, capsules...)
			if err != nil {
				t.Error(err)
			}

			count := result[0].Data()
			if !bytes.Equal(count, test.expected) {
				t.Errorf("expected %s, got %s", test.expected, count)
			}
		})
	}
}

func benchmarkCount(b *testing.B, applier procCount, capsules []config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = applier.Batch(ctx, capsules...)
	}
}

func BenchmarkCount(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range countTests {
		var capsules []config.Capsule
		for _, t := range test.test {
			capsule.SetData(t)
			capsules = append(capsules, capsule)
		}

		proc, err := newProcCount(test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkCount(b, proc, capsules)
			},
		)
	}
}
