package process

import (
	"bytes"
	"context"
	"errors"
	"testing"
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
		res, err := test.proc.Byte(ctx, test.test)
		if err != nil && errors.Is(err, test.err) {
			continue
		} else if err != nil {
			t.Log(err)
			t.Fail()
		}

		if c := bytes.Compare(res, test.expected); c != 0 {
			t.Logf("expected %s, got %s", test.expected, res)
			t.Fail()
		}
	}
}

func benchmarkConcatByte(b *testing.B, byter Concat, test []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		byter.Byte(ctx, test)
	}
}

func BenchmarkConcatByte(b *testing.B) {
	for _, test := range concatTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				benchmarkConcatByte(b, test.proc, test.test)
			},
		)
	}
}
