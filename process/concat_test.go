package process

import (
	"bytes"
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var concatTests = []struct {
	name     string
	proc     Concat
	test     []byte
	expected []byte
	err      error
}{
	{
		"JSON",
		Concat{
			Options: ConcatOptions{
				Separator: ".",
			},
			InputKey:  "foo",
			OutputKey: "foo",
		},
		[]byte(`{"foo":["bar","baz"]}`),
		[]byte(`{"foo":"bar.baz"}`),
		nil,
	},
}

func TestConcat(t *testing.T) {
	ctx := context.TODO()
	capsule := config.NewCapsule()

	for _, test := range concatTests {
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

func benchmarkConcat(b *testing.B, applicator Concat, test config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = applicator.Apply(ctx, test)
	}
}

func BenchmarkConcat(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range concatTests {
		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkConcat(b, test.proc, capsule)
			},
		)
	}
}
