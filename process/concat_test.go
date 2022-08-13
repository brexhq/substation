package process

import (
	"bytes"
	"context"
	"errors"
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
	{
		"invalid settings",
		Concat{},
		[]byte{},
		[]byte{},
		ProcessorInvalidSettings,
	},
}

func TestConcat(t *testing.T) {
	ctx := context.TODO()
	for _, test := range concatTests {

		cap := config.NewCapsule()
		cap.SetData(test.test)

		res, err := test.proc.Apply(ctx, cap)
		if err != nil && errors.Is(err, test.err) {
			continue
		} else if err != nil {
			t.Log(err)
			t.Fail()
		}

		if c := bytes.Compare(res.GetData(), test.expected); c != 0 {
			t.Logf("expected %s, got %s", test.expected, res.GetData())
			t.Fail()
		}
	}
}

func benchmarkConcatCapByte(b *testing.B, applicator Concat, test config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		applicator.Apply(ctx, test)
	}
}

func BenchmarkConcatCapByte(b *testing.B) {
	for _, test := range concatTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				cap := config.NewCapsule()
				cap.SetData(test.test)
				benchmarkConcatCapByte(b, test.proc, cap)
			},
		)
	}
}
