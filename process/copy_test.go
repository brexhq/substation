package process

import (
	"bytes"
	"context"
	"testing"
)

func TestCopy(t *testing.T) {
	var tests = []struct {
		proc     Copy
		test     []byte
		expected []byte
	}{
		{
			Copy{
				Input: Input{
					Key: "original",
				},
				Output: Output{
					Key: "copy",
				},
			},
			[]byte(`{"original":"hello"}`),
			[]byte(`{"original":"hello","copy":"hello"}`),
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
