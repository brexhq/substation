package process

import (
	"bytes"
	"context"
	"testing"

	"github.com/brexhq/substation/condition"
)

func TestExpand(t *testing.T) {
	var tests = []struct {
		proc     Expand
		test     []byte
		expected [][]byte
	}{
		{
			Expand{
				Input: Input{
					Key: "Expand",
				},
				// Options: Output{
				// 	Key: "Expand",
				// },
			},
			[]byte(`{"expand":["123","456"]}`),
			[][]byte{
				[]byte(`{"expand":"123"}`),
				[]byte(`{"expand":"456"}`),
			},
		},
		{
			Expand{
				Condition: condition.OperatorConfig{
					Operator: "all",
				},
				Input: Input{
					Key: "Expand",
				},
				Options: ExpandOptions{
					Retain: []string{"foo"},
				},
			},
			[]byte(`{"expand":["123","456"],"foo":"bar"}`),
			[][]byte{
				[]byte(`{"expand":"123","foo":"bar"}`),
				[]byte(`{"expand":"456","foo":"bar"}`),
			},
		},
		{
			Expand{
				Condition: condition.OperatorConfig{
					Operator: "all",
				},
				Input: Input{
					Key: "Expand",
				},
				Options: ExpandOptions{
					Retain: []string{"foo", "baz"},
				},
			},
			[]byte(`{"expand":["123","456"],"foo":"bar","baz":"qux"}`),
			[][]byte{
				[]byte(`{"expand":"123","foo":"bar","baz":"qux"}`),
				[]byte(`{"expand":"456","foo":"bar","baz":"qux"}`),
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
		}
	}
}
