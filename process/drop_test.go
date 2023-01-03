package process

import (
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var dropTests = []struct {
	name string
	proc procDrop
	test [][]byte
	err  error
}{
	{
		"drop",
		procDrop{},
		[][]byte{
			[]byte(`{"foo":"bar"}`),
			[]byte(`{"foo":"baz"}`),
			[]byte(`{"foo":"qux"}`),
		},
		nil,
	},
}

func TestDrop(t *testing.T) {
	ctx := context.TODO()
	capsule := config.NewCapsule()

	for _, test := range dropTests {
		var capsules []config.Capsule
		for _, t := range test.test {
			capsule.SetData(t)
			capsules = append(capsules, capsule)
		}

		result, err := test.proc.Batch(ctx, capsules...)
		if err != nil {
			t.Error(err)
		}

		length := len(result)
		if length != 0 {
			t.Errorf("got %d", length)
		}
	}
}

func benchmarkDrop(b *testing.B, applier procDrop, capsules []config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = applier.Batch(ctx, capsules...)
	}
}

func BenchmarkDrop(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range dropTests {
		var _ Batcher = test.proc

		var capsules []config.Capsule
		for _, t := range test.test {
			capsule.SetData(t)
			capsules = append(capsules, capsule)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkDrop(b, test.proc, capsules)
			},
		)
	}
}
