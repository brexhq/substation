package process

import (
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var dropTests = []struct {
	name string
	proc Drop
	test [][]byte
	err  error
}{
	{
		"drop",
		Drop{},
		[][]byte{
			[]byte(`{"foo":"bar"}`),
			[]byte(`{"foo":"baz"}`),
			[]byte(`{"foo":"qux"}`),
		},
		nil,
	},
}

func TestDrop(t *testing.T) {
	ctx := context.TODO()
	cap := config.NewCapsule()

	for _, test := range dropTests {
		var caps []config.Capsule
		for _, t := range test.test {
			cap.SetData(t)
			caps = append(caps, cap)
		}

		result, err := test.proc.ApplyBatch(ctx, caps)
		if err != nil {
			t.Log(err)
			t.Fail()
		}

		length := len(result)
		if length != 0 {
			t.Logf("got %d", length)
			t.Fail()
		}
	}
}

func benchmarkDrop(b *testing.B, applicator Drop, caps []config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		applicator.ApplyBatch(ctx, caps)
	}
}

func BenchmarkDrop(b *testing.B) {
	cap := config.NewCapsule()
	for _, test := range dropTests {
		var caps []config.Capsule
		for _, t := range test.test {
			cap.SetData(t)
			caps = append(caps, cap)
		}

		b.Run(string(test.name),
			func(b *testing.B) {
				benchmarkDrop(b, test.proc, caps)
			},
		)
	}
}
