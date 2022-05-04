package process

import (
	"context"
	"testing"
)

var dropTests = []struct {
	name string
	proc Drop
	test [][]byte
}{
	{
		"drop",
		Drop{},
		[][]byte{
			[]byte(`{"foo":"bar"}`),
			[]byte(`{"foo":"baz"}`),
			[]byte(`{"foo":"qux"}`),
		},
	},
}

func TestDrop(t *testing.T) {
	ctx := context.TODO()
	for _, test := range dropTests {
		pipe := make(chan []byte, len(test.test))
		for _, x := range test.test {
			pipe <- x
		}
		close(pipe)

		res, err := test.proc.Channel(ctx, pipe)
		if err != nil {
			t.Log(err)
			t.Fail()
		}

		if len(res) != 0 {
			t.Log("result pipe wrong size")
			t.Fail()
		}
	}
}

func benchmarkDropByte(b *testing.B, byter Drop, test chan []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		byter.Channel(ctx, test)
	}
}

func BenchmarkDropByte(b *testing.B) {
	for _, test := range dropTests {
		pipe := make(chan []byte, len(test.test))
		for _, x := range test.test {
			pipe <- x
		}
		close(pipe)

		b.Run(string(test.name),
			func(b *testing.B) {
				benchmarkDropByte(b, test.proc, pipe)
			},
		)
	}
}
