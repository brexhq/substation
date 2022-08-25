package process

import (
	"bytes"
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var flattenTests = []struct {
	name     string
	proc     Flatten
	test     []byte
	expected []byte
	err      error
}{
	{
		"json",
		Flatten{
			InputKey:  "flatten",
			OutputKey: "flatten",
		},
		[]byte(`{"flatten":["foo",["bar"]]}`),
		[]byte(`{"flatten":["foo","bar"]}`),
		nil,
	},
	{
		"json deep flatten",
		Flatten{
			Options: FlattenOptions{
				Deep: true,
			},
			InputKey:  "flatten",
			OutputKey: "flatten",
		},
		[]byte(`{"flatten":[["foo"],[[["bar",[["baz"]]]]]]}`),
		[]byte(`{"flatten":["foo","bar","baz"]}`),
		nil,
	},
}

func TestFlatten(t *testing.T) {
	ctx := context.TODO()
	cap := config.NewCapsule()

	for _, test := range flattenTests {
		cap.SetData(test.test)

		result, err := test.proc.Apply(ctx, cap)
		if err != nil {
			t.Log(err)
			t.Fail()
		}

		if !bytes.Equal(result.GetData(), test.expected) {
			t.Logf("expected %s, got %s", test.expected, result.GetData())
			t.Fail()
		}
	}
}

func benchmarkFlatten(b *testing.B, applicator Flatten, test config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		applicator.Apply(ctx, test)
	}
}

func BenchmarkFlatten(b *testing.B) {
	cap := config.NewCapsule()
	for _, test := range flattenTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				cap.SetData(test.test)
				benchmarkFlatten(b, test.proc, cap)
			},
		)
	}
}
