package process

import (
	"bytes"
	"context"
	"errors"
	"testing"
)

var mathTests = []struct {
	name     string
	proc     Math
	err      error
	test     []byte
	expected []byte
}{
	{
		"add",
		Math{
			Options: MathOptions{
				Operation: "add",
			},
			InputKey:  "math",
			OutputKey: "math",
		},
		nil,
		[]byte(`{"math":[1,3]}`),
		[]byte(`{"math":4}`),
	},
	{
		"subtract",
		Math{
			Options: MathOptions{
				Operation: "subtract",
			},
			InputKey:  "math",
			OutputKey: "math",
		},
		nil,
		[]byte(`{"math":[5,2]}`),
		[]byte(`{"math":3}`),
	},
	{
		"divide",
		Math{
			Options: MathOptions{
				Operation: "divide",
			},
			InputKey:  "math",
			OutputKey: "math",
		},
		nil,
		[]byte(`{"math":[10,2]}`),
		[]byte(`{"math":5}`),
	},
}

func TestMath(t *testing.T) {
	for _, test := range mathTests {
		ctx := context.TODO()
		res, err := test.proc.Byte(ctx, test.test)
		if err != nil && errors.As(err, &test.err) {
			continue
		} else if err != nil {
			t.Log(err)
			t.Fail()
		}

		if c := bytes.Compare(res, test.expected); c != 0 {
			t.Logf("expected %s, got %s", test.expected, res)
			t.Fail()
		}
	}
}

func benchmarkMathByte(b *testing.B, byter Math, test []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		byter.Byte(ctx, test)
	}
}

func BenchmarkMathByte(b *testing.B) {
	for _, test := range mathTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				benchmarkMathByte(b, test.proc, test.test)
			},
		)
	}
}
