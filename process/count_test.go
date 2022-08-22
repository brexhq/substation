package process

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/brexhq/substation/config"
)

var countTests = []struct {
	name     string
	proc     Count
	test     [][]byte
	expected []byte
	err      error
}{
	{
		"count",
		Count{},
		[][]byte{
			[]byte(`{"foo":"bar"}`),
			[]byte(`{"foo":"baz"}`),
			[]byte(`{"foo":"qux"}`),
		},
		[]byte(`{"count":3}`),
		nil,
	},
}

func TestCount(t *testing.T) {
	ctx := context.TODO()
	cap := config.NewCapsule()
	for _, test := range countTests {
		var caps []config.Capsule
		for _, t := range test.test {
			cap.SetData(t)
			caps = append(caps, cap)
		}

		res, err := test.proc.ApplyBatch(ctx, caps)
		if err != nil && errors.Is(err, test.err) {
			continue
		} else if err != nil {
			t.Log(err)
			t.Fail()
		}

		count := res[0].GetData()
		if !bytes.Equal(count, test.expected) {
			t.Logf("expected %s, got %s", test.expected, count)
			t.Fail()
		}
	}
}

func benchmarkCount(b *testing.B, applicator Count, caps []config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		applicator.ApplyBatch(ctx, caps)
	}
}

func BenchmarkCount(b *testing.B) {
	cap := config.NewCapsule()
	for _, test := range countTests {
		var caps []config.Capsule
		for _, t := range test.test {
			cap.SetData(t)
			caps = append(caps, cap)
		}

		b.Run(string(test.name),
			func(b *testing.B) {
				benchmarkCount(b, test.proc, caps)
			},
		)
	}
}
