package process

import (
	"bytes"
	"context"
	"testing"
)

func TestFlatten(t *testing.T) {
	var tests = []struct {
		proc     Flatten
		test     []byte
		expected []byte
	}{
		{
			Flatten{
				Input: Input{
					Key: "flatten",
				},
				Output: Output{
					Key: "flatten",
				},
			},
			[]byte(`{"flatten":["123",["456"]]}`),
			[]byte(`{"flatten":["123","456"]}`),
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
