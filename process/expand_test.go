package process

import (
	"bytes"
	"context"
	"testing"
)

var expandTests = []struct {
	name     string
	proc     Expand
	test     []byte
	expected [][]byte
}{
	{
		"json",
		Expand{
			Input: Input{
				Key: "expand",
			},
		},
		[]byte(`{"expand":[{"foo":"bar"}],"baz":"qux"`),
		[][]byte{
			[]byte(`{"foo":"bar"}`),
		},
	},
	{
		"json retain",
		Expand{
			Input: Input{
				Key: "expand",
			},
			Options: ExpandOptions{
				Retain: []string{"baz"},
			},
		},
		[]byte(`{"expand":[{"foo":"bar"}],"baz":"qux"`),
		[][]byte{
			[]byte(`{"foo":"bar","baz":"qux"}`),
		},
	},
}

func TestExpand(t *testing.T) {
	ctx := context.TODO()
	for _, test := range expandTests {
		slice := make([][]byte, 1, 1)
		slice[0] = test.test

		res, err := test.proc.Slice(ctx, slice)
		if err != nil {
			t.Log(err)
			t.Fail()
		}

		count := 0
		for _, processed := range res {
			expected := test.expected[count]
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
