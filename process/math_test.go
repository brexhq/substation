package process

import (
	"bytes"
	"context"
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

func benchmarkMath(b *testing.B, applicator Math, test config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = applicator.Apply(ctx, test)
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
