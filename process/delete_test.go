package process

import (
	"bytes"
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var deleteTests = []struct {
	name     string
	proc     _delete
	test     []byte
	expected []byte
	err      error
}{
	{
		"string",
		_delete{
			process: process{
				Key: "baz",
			},
		},
		[]byte(`{"foo":"bar","baz":"qux"}`),
		[]byte(`{"foo":"bar"}`),
		nil,
	},
	{
		"JSON",
		_delete{
			process: process{
				Key: "baz",
			},
		},
		[]byte(`{"foo":"bar","baz":{"qux":"quux"}}`),
		[]byte(`{"foo":"bar"}`),
		nil,
	},
}

func TestDelete(t *testing.T) {
	ctx := context.TODO()
	capsule := config.NewCapsule()

	for _, test := range deleteTests {
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

func benchmarkDelete(b *testing.B, applier _delete, test config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = applier.Apply(ctx, test)
	}
}

func BenchmarkDelete(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range deleteTests {
		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkDelete(b, test.proc, capsule)
			},
		)
	}
}
