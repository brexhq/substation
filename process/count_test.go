package process

import (
	"bytes"
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var countTests = []struct {
	name     string
	proc     count
	test     [][]byte
	expected []byte
	err      error
}{
	{
		"count",
		count{},
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
		var capsules []config.Capsule
		for _, t := range test.test {
			capsule.SetData(t)
			capsules = append(capsules, capsule)
		}

		result, err := test.proc.Batch(ctx, capsules...)
		if err != nil {
			t.Error(err)
		}

		count := result[0].Data()
		if !bytes.Equal(count, test.expected) {
			t.Errorf("expected %s, got %s", test.expected, count)
		}
	}
}

func benchmarkCount(b *testing.B, applicator count, capsules []config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = applicator.Batch(ctx, capsules...)
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

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkCount(b, test.proc, capsules)
			},
		)
	}
}
