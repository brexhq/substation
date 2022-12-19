package process

import (
	"bytes"
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var mathTests = []struct {
	name     string
	proc     math
	test     []byte
	expected []byte
	err      error
}{
	{
		"add",
		math{
			process: process{
				Key:    "foo",
				SetKey: "foo",
			},
			Options: mathOptions{
				Operation: "add",
			},
		},
		[]byte(`{"foo":[1,3]}`),
		[]byte(`{"foo":4}`),
		nil,
	},
	{
		"subtract",
		math{
			process: process{
				Key:    "foo",
				SetKey: "foo",
			},
			Options: mathOptions{
				Operation: "subtract",
			},
		},
		[]byte(`{"foo":[5,2]}`),
		[]byte(`{"foo":3}`),
		nil,
	},
	{
		"multiply",
		math{
			process: process{
				Key:    "foo",
				SetKey: "foo",
			},
			Options: mathOptions{
				Operation: "multiply",
			},
		},
		[]byte(`{"foo":[10,2]}`),
		[]byte(`{"foo":20}`),
		nil,
	},
	{
		"divide",
		math{
			process: process{
				Key:    "foo",
				SetKey: "foo",
			},
			Options: mathOptions{
				Operation: "divide",
			},
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

func benchmarkMath(b *testing.B, applicator math, test config.Capsule) {
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
