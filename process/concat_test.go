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
	cap := config.NewCapsule()

	for _, test := range concatTests {
		cap.SetData(test.test)

		result, err := test.proc.Apply(ctx, cap)
		if err != nil {
			t.Log(err)
			t.Fail()
		}

		if !bytes.Equal(result.Data(), test.expected) {
			t.Logf("expected %s, got %s", test.expected, result.Data())
			t.Fail()
		}
	}
}

func benchmarkConcat(b *testing.B, applicator Concat, test config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		applicator.Apply(ctx, test)
	}
}

func BenchmarkConcat(b *testing.B) {
	cap := config.NewCapsule()
	for _, test := range concatTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				cap.SetData(test.test)
				benchmarkConcat(b, test.proc, cap)
			},
		)
	}
}
