package process

import (
	"bytes"
	"context"
	"testing"
)

func TestCase(t *testing.T) {
	var tests = []struct {
		proc     Case
		test     []byte
		expected []byte
	}{
		{
			Case{
				Input: Input{
					Key: "case",
				},
				Options: CaseOptions{
					Case: "lower",
				},
				Output: Output{
					Key: "case",
				},
			},
			[]byte(`{"case":"ABC"}`),
			[]byte(`{"case":"abc"}`),
		},
		{
			Case{
				Input: Input{
					Key: "case",
				},
				Options: CaseOptions{
					Case: "upper",
				},
				Output: Output{
					Key: "case",
				},
			},
			[]byte(`{"case":"abc"}`),
			[]byte(`{"case":"ABC"}`),
		},
		{
			Case{
				Input: Input{
					Key: "case",
				},
				Options: CaseOptions{
					Case: "snake",
				},
				Output: Output{
					Key: "case",
				},
			},
			[]byte(`{"case":"AbC"})`),
			[]byte(`{"case":"ab_c"})`),
		},
		// array support
		{
			Case{
				Input: Input{
					Key: "case",
				},
				Options: CaseOptions{
					Case: "lower",
				},
				Output: Output{
					Key: "case",
				},
			},
			[]byte(`{"case":["ABC","DEF"]}`),
			[]byte(`{"case":["abc","def"]}`),
		},
		{
			Case{
				Input: Input{
					Key: "case",
				},
				Options: CaseOptions{
					Case: "upper",
				},
				Output: Output{
					Key: "case",
				},
			},
			[]byte(`{"case":["abc","def"]}`),
			[]byte(`{"case":["ABC","DEF"]}`),
		},
		{
			Case{
				Input: Input{
					Key: "case",
				},
				Options: CaseOptions{
					Case: "snake",
				},
				Output: Output{
					Key: "case",
				},
			},
			[]byte(`{"case":["AbC","DeF"]}`),
			[]byte(`{"case":["ab_c","de_f"]}`),
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
