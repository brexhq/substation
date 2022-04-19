package process

import (
	"bytes"
	"context"
	"testing"
)

func TestCount(t *testing.T) {
	var tests = []struct {
		proc     Count
		test     [][]byte
		expected []byte
	}{
		{
			Count{},
			[][]byte{
				[]byte(`{"foo":"bar"}`),
				[]byte(`{"foo":"baz"}`),
				[]byte(`{"foo":"qux"}`),
			},
			[]byte(`{"count":3}`),
		},
	}

	ctx := context.TODO()

	for _, test := range tests {
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
