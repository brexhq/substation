package process

import (
	"bytes"
	"context"
	"errors"
	"testing"
)

var countTests = []struct {
	name     string
	proc     Count
	test     [][]byte
	expected []byte
	err      error
}{
	{
		"count",
		Count{},
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
	for _, test := range countTests {
		res, err := test.proc.Slice(ctx, test.test)
		if err != nil && errors.As(err, &test.err) {
			continue
		} else if err != nil {
			t.Log(err)
			t.Fail()
		}

		if len(res) != 1 {
			t.Log("result pipe wrong size")
		}

		if c := bytes.Compare(test.expected, res[0]); c != 0 {
			t.Logf("expected %s, got %s", test.expected, res[0])
			t.Fail()
		}
	}
}

func benchmarkCountSlice(b *testing.B, slicer Count, test [][]byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		slicer.Slice(ctx, test)
	}
}

func BenchmarkCountSlice(b *testing.B) {
	for _, test := range countTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				benchmarkCountSlice(b, test.proc, test.test)
			},
		)
	}
}
