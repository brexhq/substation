package process

import (
	"bytes"
	"context"
	"errors"
	"testing"
)

var expandTests = []struct {
	name     string
	proc     Expand
	test     []byte
	expected [][]byte
	err      error
}{
	{
		"JSON",
		Expand{
			InputKey: "expand",
		},
		[]byte(`{"expand":[{"foo":"bar"}]}`),
		[][]byte{
			[]byte(`{"foo":"bar"}`),
		},
		nil,
	},
	{
		"JSON extra key",
		Expand{
			InputKey: "expand",
		},
		[]byte(`{"expand":[{"foo":"bar"},{"quux":"corge"}],"baz":"qux"}`),
		[][]byte{
			[]byte(`{"foo":"bar","baz":"qux"}`),
			[]byte(`{"quux":"corge","baz":"qux"}`),
		},
		nil,
	},
	{
		"invalid settings",
		Expand{},
		[]byte{},
		[][]byte{},
		ProcessorInvalidSettings,
	},
}

func TestExpand(t *testing.T) {
	ctx := context.TODO()
	for _, test := range expandTests {
		slice := make([][]byte, 1, 1)
		slice[0] = test.test

		res, err := test.proc.Slice(ctx, slice)
		if err != nil && errors.Is(err, test.err) {
			continue
		} else if err != nil {
			t.Log(err)
			t.Fail()
		}

		for i, processed := range res {
			expected := test.expected[i]
			if c := bytes.Compare(expected, processed); c != 0 {
				t.Logf("expected %s, got %s", expected, processed)
				t.Fail()
			}
		}
	}
}

func benchmarkExpandSlice(b *testing.B, slicer Expand, slice [][]byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		slicer.Slice(ctx, slice)
	}
}

func BenchmarkExpandSlice(b *testing.B) {
	for _, test := range expandTests {
		slice := make([][]byte, 1, 1)
		slice[0] = test.test

		b.Run(string(test.name),
			func(b *testing.B) {
				benchmarkExpandSlice(b, test.proc, slice)
			},
		)
	}
}
