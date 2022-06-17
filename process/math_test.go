package process

import (
	"bytes"
	"context"
	"testing"
)

var mathTests = []struct {
	name     string
	proc     Math
	test     []byte
	expected []byte
}{
	{
		"add",
		Math{
			InputKey:  "math",
			OutputKey: "math",
			Options: MathOptions{
				Operation: "add",
			},
		},
		[]byte(`{"math":[1,3]}`),
		[]byte(`{"math":4}`),
	},
	{
		"subtract",
		Math{
			InputKey:  "math",
			OutputKey: "math",
			Options: MathOptions{
				Operation: "subtract",
			},
		},
		[]byte(`{"math":[5,2]}`),
		[]byte(`{"math":3}`),
	},
	{
		"divide",
		Math{
			InputKey:  "math",
			OutputKey: "math",
			Options: MathOptions{
				Operation: "divide",
			},
		},
		[]byte(`{"math":[10,2]}`),
		[]byte(`{"math":5}`),
	},
	{
		"add array",
		Math{
			InputKey:  "math",
			OutputKey: "math",
			Options: MathOptions{
				Operation: "add",
			},
		},
		[]byte(`{"math":[[1,2],[3,4]]}`),
		[]byte(`{"math":[4,6]}`),
	},
	{
		"subtract array",
		Math{
			InputKey:  "math",
			OutputKey: "math",
			Options: MathOptions{
				Operation: "subtract",
			},
		},
		[]byte(`{"math":[[10,5],[4,1]]}`),
		[]byte(`{"math":[6,4]}`),
	},
	{
		"divide array",
		Math{
			InputKey:  "math",
			OutputKey: "math",
			Options: MathOptions{
				Operation: "divide",
			},
		},
		[]byte(`{"math":[[10,5],[5,1]]}`),
		[]byte(`{"math":[2,5]}`),
	},
}

func TestMath(t *testing.T) {
	for _, test := range mathTests {
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
