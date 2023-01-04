package process

import (
	"bytes"
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var joinTests = []struct {
	name     string
	proc     procJoin
	test     []byte
	expected []byte
	err      error
}{
	{
		"JSON",
		procJoin{
			process: process{
				Key:    "foo",
				SetKey: "foo",
			},
			Options: procJoinOptions{
				Separator: ".",
			},
		},
		[]byte(`{"foo":["bar","baz"]}`),
		[]byte(`{"foo":"bar.baz"}`),
		nil,
	},
}

func TestJoin(t *testing.T) {
	ctx := context.TODO()
	capsule := config.NewCapsule()

	for _, test := range joinTests {
		var _ Applier = test.proc
		var _ Batcher = test.proc

		capsule.SetData(test.test)

		result, err := test.proc.Apply(ctx, capsule)
		if err != nil {
			t.Error(err)
		}

		if !bytes.Equal(result.Data(), test.expected) {
			t.Errorf("expected %s, got %s", test.expected, result.Data())
		}
	}
}

func benchmarkJoin(b *testing.B, applier procJoin, test config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = applier.Apply(ctx, test)
	}
}

func BenchmarkJoin(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range joinTests {
		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkJoin(b, test.proc, capsule)
			},
		)
	}
}
