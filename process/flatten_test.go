package process

import (
	"bytes"
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var flattenTests = []struct {
	name     string
	proc     flatten
	test     []byte
	expected []byte
	err      error
}{
	{
		"json",
		flatten{
			process: process{
				Key:    "flatten",
				SetKey: "flatten",
			},
		},
		[]byte(`{"flatten":["foo",["bar"]]}`),
		[]byte(`{"flatten":["foo","bar"]}`),
		nil,
	},
	{
		"json deep flatten",
		flatten{
			process: process{
				Key:    "flatten",
				SetKey: "flatten",
			},
			Options: flattenOptions{
				Deep: true,
			},
		},
		[]byte(`{"flatten":[["foo"],[[["bar",[["baz"]]]]]]}`),
		[]byte(`{"flatten":["foo","bar","baz"]}`),
		nil,
	},
}

func TestFlatten(t *testing.T) {
	ctx := context.TODO()
	capsule := config.NewCapsule()

	for _, test := range flattenTests {
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

func benchmarkFlatten(b *testing.B, applicator flatten, test config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = applicator.Apply(ctx, test)
	}
}

func BenchmarkFlatten(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range flattenTests {
		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkFlatten(b, test.proc, capsule)
			},
		)
	}
}
