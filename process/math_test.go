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
	test     []byte
	expected []byte
	err      error
}{
	{
		"add",
		Math{
			Options: MathOptions{
				Operation: "add",
			},
			InputKey:  "foo",
			OutputKey: "foo",
		},
		[]byte(`{"foo":[1,3]}`),
		[]byte(`{"foo":4}`),
		nil,
	},
	{
		"subtract",
		Math{
			Options: MathOptions{
				Operation: "subtract",
			},
			InputKey:  "foo",
			OutputKey: "foo",
		},
		[]byte(`{"foo":[5,2]}`),
		[]byte(`{"foo":3}`),
		nil,
	},
	{
		"multiply",
		Math{
			Options: MathOptions{
				Operation: "multiply",
			},
			InputKey:  "foo",
			OutputKey: "foo",
		},
		[]byte(`{"foo":[10,2]}`),
		[]byte(`{"foo":20}`),
		nil,
	},
	{
		"divide",
		Math{
			Options: MathOptions{
				Operation: "divide",
			},
			InputKey:  "foo",
			OutputKey: "foo",
		},
		[]byte(`{"foo":[10,2]}`),
		[]byte(`{"foo":5}`),
		nil,
	},
	{
		"invalid settings",
		Math{},
		[]byte{},
		[]byte{},
		ProcessorInvalidSettings,
	},
}

func TestMath(t *testing.T) {
	for _, test := range mathTests {
		ctx := context.TODO()
		res, err := test.proc.Byte(ctx, test.test)
		if err != nil && errors.Is(err, test.err) {
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
