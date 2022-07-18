package process

import (
	"context"
	"errors"
	"testing"
)

var dropTests = []struct {
	name string
	proc Drop
	test [][]byte
	err  error
}{
	{
		"drop",
		Drop{},
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
	for _, test := range dropTests {
		res, err := test.proc.Slice(ctx, test.test)
		if err != nil && errors.Is(err, test.err) {
			continue
		} else if err != nil {
			t.Log(err)
			t.Fail()
		}

		if len(res) != 0 {
			t.Log("result pipe wrong size")
			t.Fail()
		}
	}
}

func benchmarkDropSlice(b *testing.B, slicer Drop, test [][]byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		slicer.Slice(ctx, test)
	}
}

func BenchmarkDropSlice(b *testing.B) {
	for _, test := range dropTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				benchmarkDropSlice(b, test.proc, test.test)
			},
		)
	}
}
