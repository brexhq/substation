package process

import (
	"context"
	"testing"
)

func TestDrop(t *testing.T) {
	var tests = []struct {
		proc Drop
		test [][]byte
	}{
		{
			Drop{},
			[][]byte{
				[]byte(`{"foo":"bar"}`),
				[]byte(`{"foo":"baz"}`),
				[]byte(`{"foo":"qux"}`),
			},
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

		if len(res) != 0 {
			t.Log("result pipe wrong size")
			t.Fail()
		}
	}
}
