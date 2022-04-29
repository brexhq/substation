package process

import (
	"bytes"
	"context"
	"testing"
)

func TestSplitChannel(t *testing.T) {
	var tests = []struct {
		proc     Split
		test     []byte
		expected [][]byte
	}{
		{
			Split{
				Options: SplitOptions{
					Separator: `\n`,
				},
			},
			[]byte(`{"hello":123}\n{"hello":456}\n{"hello":789}`),
			[][]byte{
				[]byte(`{"hello":123}`),
				[]byte(`{"hello":456}`),
				[]byte(`{"hello":789}`),
			},
		},
	}

	ctx := context.TODO()

	for _, test := range tests {
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
			count++
		}
	}
}

func TestSplitByte(t *testing.T) {
	var tests = []struct {
		proc     Split
		test     []byte
		expected []byte
	}{
		{
			Split{
				Input: Input{
					Key: "hello",
				},
				Options: SplitOptions{
					Separator: `,`,
				},
				Output: Output{
					Key: "hello",
				},
			},
			[]byte(`{"hello":"123,456,789"}`),
			[]byte(`{"hello":["123","456","789"]}`),
		},
		{
			Split{
				Input: Input{
					Key: "hello",
				},
				Options: SplitOptions{
					Separator: `,`,
				},
				Output: Output{
					Key: "hello",
				},
			},
			[]byte(`{"hello":["12,34","56,78"]}`),
			[]byte(`{"hello":[["12","34"],["56","78"]]}`),
		},
	}

	for _, test := range tests {
		ctx := context.TODO()
		res, err := test.proc.Byte(ctx, test.test)
		if err != nil {
			t.Logf("%v", err)
			t.Fail()
		}

		if c := bytes.Compare(res, test.expected); c != 0 {
			t.Logf("expected %s, got %s", test.expected, res)
			t.Fail()
		}
	}
}
