package process

import (
	"bytes"
	"context"
	"testing"
)

var countTests = []struct {
	name     string
	proc     Count
	test     [][]byte
	expected []byte
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
	},
}

func TestCount(t *testing.T) {

	ctx := context.TODO()

	for _, test := range countTests {
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

		if len(res) != 1 {
			t.Log("result pipe wrong size")
		}

		processed := <-res
		if c := bytes.Compare(test.expected, processed); c != 0 {
			t.Logf("expected %s, got %s", test.expected, processed)
			t.Fail()
		}
	}
}

func benchmarkCountByte(b *testing.B, byter Count, test chan []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		byter.Channel(ctx, test)
	}
}

func BenchmarkCountByte(b *testing.B) {
	for _, test := range countTests {
		pipe := make(chan []byte, len(test.test))
		for _, x := range test.test {
			pipe <- x
		}
		close(pipe)

		b.Run(string(test.name),
			func(b *testing.B) {
				benchmarkCountByte(b, test.proc, pipe)
			},
		)
	}
}
