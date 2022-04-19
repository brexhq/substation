package process

import (
	"bytes"
	"context"
	"testing"
)

func TestReplace(t *testing.T) {
	var tests = []struct {
		proc     Replace
		test     []byte
		expected []byte
	}{
		{
			Replace{
				Input: Input{
					Key: "replace",
				},
				Options: ReplaceOptions{
					Old: "l",
					New: "|",
				},
				Output: Output{
					Key: "replace",
				},
			},
			[]byte(`{"replace":"hello"}`),
			[]byte(`{"replace":"he||o"}`),
		},
		// array input
		{
			Replace{
				Input: Input{
					Key: "replace",
				},
				Options: ReplaceOptions{
					Old: "l",
					New: "|",
				},
				Output: Output{
					Key: "replace",
				},
			},
			[]byte(`{"replace":["hello","halo"]}`),
			[]byte(`{"replace":["he||o","ha|o"]}`),
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
