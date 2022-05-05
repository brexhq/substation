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
					Keys: []string{"foo", "bar"},
				},
				Options: MathOptions{
					Operation: "add",
				},
				Output: Output{
					Key: "math",
				},
			},
			[]byte(`{"foo":1,"bar":3}`),
			[]byte(`{"foo":1,"bar":3,"math":4}`),
		},
		{
			Math{
				Input: Inputs{
					Keys: []string{"foo", "bar"},
				},
				Options: MathOptions{
					Operation: "subtract",
				},
				Output: Output{
					Key: "math",
				},
			},
			[]byte(`{"foo":5,"bar":2}`),
			[]byte(`{"foo":5,"bar":2,"math":3}`),
		},
		{
			Math{
				Input: Inputs{
					Keys: []string{"foo", "bar"},
				},
				Options: MathOptions{
					Operation: "add",
				},
				Output: Output{
					Key: "math",
				},
			},
			[]byte(`{"foo":[1,2],"bar":[3,4]}`),
			[]byte(`{"foo":[1,2],"bar":[3,4],"math":[4,6]}`),
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
