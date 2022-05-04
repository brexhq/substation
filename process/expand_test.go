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
		pipe := make(chan []byte, 1)
		pipe <- test.test
		close(pipe)

		res, err := test.proc.Channel(ctx, pipe)
		if err != nil {
			t.Log(err)
			t.Fail()
		}

		count := 0
		for processed := range res {
			expected := test.expected[count]
			if c := bytes.Compare(expected, processed); c != 0 {
				t.Logf("expected %s, got %s", expected, processed)
				t.Fail()
			}
		}
	}
}

func benchmarkExpandByte(b *testing.B, byter Expand, test chan []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		byter.Channel(ctx, test)
	}
}

func BenchmarkExpandByte(b *testing.B) {
	for _, test := range expandTests {
		pipe := make(chan []byte, len(test.test))
		pipe <- test.test
		close(pipe)

		b.Run(string(test.name),
			func(b *testing.B) {
				benchmarkExpandByte(b, test.proc, pipe)
			},
		)
	}
}
