package process

import (
	"bytes"
	"context"
	"testing"
)

func TestMath(t *testing.T) {
	var tests = []struct {
		proc     Math
		test     []byte
		expected []byte
	}{
		{
			Math{
				Input: Inputs{
					Keys: []string{"first", "second"},
				},
				Options: MathOptions{
					Operation: "add",
				},
				Output: Output{
					Key: "third",
				},
			},
			[]byte(`{"first":5,"second":2}`),
			[]byte(`{"first":5,"second":2,"third":7}`),
		},
		{
			Math{
				Input: Inputs{
					Keys: []string{"first", "second"},
				},
				Options: MathOptions{
					Operation: "subtract",
				},
				Output: Output{
					Key: "third",
				},
			},
			[]byte(`{"first":5,"second":2}`),
			[]byte(`{"first":5,"second":2,"third":3}`),
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
