package process

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/brexhq/substation/config"
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
	ctx := context.TODO()
	for _, test := range mathTests {

		cap := config.NewCapsule()
		cap.SetData(test.test)

		res, err := test.proc.Apply(ctx, cap)
		if err != nil && errors.Is(err, test.err) {
			continue
		} else if err != nil {
			t.Log(err)
			t.Fail()
		}

		if c := bytes.Compare(res.GetData(), test.expected); c != 0 {
			t.Logf("expected %s, got %s", test.expected, res.GetData())
			t.Fail()
		}
	}
}

func benchmarkMathCapByte(b *testing.B, applicator Math, test config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		applicator.Apply(ctx, test)
	}
}

func BenchmarkMathCapByte(b *testing.B) {
	for _, test := range mathTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				cap := config.NewCapsule()
				cap.SetData(test.test)
				benchmarkMathCapByte(b, test.proc, cap)
			},
		)
	}
}
