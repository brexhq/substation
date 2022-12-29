package process

import (
	"bytes"
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var mathTests = []struct {
	name     string
	proc     _math
	test     []byte
	expected []byte
	err      error
}{
	{
		"add",
		_math{
			process: process{
				Key:    "math",
				SetKey: "math",
			},
			Options: _mathOptions{
				Operation: "add",
			},
		},
		[]byte(`{"math":[1,3]}`),
		[]byte(`{"math":4}`),
		nil,
	},
	{
		"subtract",
		_math{
			process: process{
				Key:    "math",
				SetKey: "math",
			},
			Options: _mathOptions{
				Operation: "subtract",
			},
		},
		[]byte(`{"math":[5,2]}`),
		[]byte(`{"math":3}`),
		nil,
	},
	{
		"multiply",
		_math{
			process: process{
				Key:    "math",
				SetKey: "math",
			},
			Options: _mathOptions{
				Operation: "multiply",
			},
		},
		[]byte(`{"math":[10,2]}`),
		[]byte(`{"math":20}`),
		nil,
	},
	{
		"divide",
		_math{
			process: process{
				Key:    "math",
				SetKey: "math",
			},
			Options: _mathOptions{
				Operation: "divide",
			},
		},
		[]byte(`{"math":[10,2]}`),
		[]byte(`{"math":5}`),
		nil,
	},
}

func TestMath(t *testing.T) {
	ctx := context.TODO()
	capsule := config.NewCapsule()

	for _, test := range mathTests {
		capsule.SetData(test.test)

		result, err := test.proc.Apply(ctx, capsule)
		if err != nil {
			t.Error(err)
		}

		if !bytes.Equal(result.Data(), test.expected) {
			t.Errorf("expected %s, got %s", test.expected, result.Data())
		}
	}
}

func benchmarkMath(b *testing.B, applier _math, test config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = applier.Apply(ctx, test)
	}
}

func BenchmarkMath(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range mathTests {
		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkMath(b, test.proc, capsule)
			},
		)
	}
}
