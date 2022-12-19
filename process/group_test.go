package process

import (
	"bytes"
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var groupTests = []struct {
	name     string
	proc     group
	test     []byte
	expected []byte
	err      error
}{
	{
		"tuples",
		group{
			process: process{
				Key:    "foo",
				SetKey: "foo",
			},
		},
		[]byte(`{"foo":[["bar","baz"],[123,456]]}`),
		[]byte(`{"foo":[["bar",123],["baz",456]]}`),
		nil,
	},
	{
		"objects",
		group{
			process: process{
				Key:    "foo",
				SetKey: "foo",
			},
			Options: groupOptions{
				Keys: []string{"qux.quux", "corge"},
			},
		},
		[]byte(`{"foo":[["bar","baz"],[123,456]]}`),
		[]byte(`{"foo":[{"qux":{"quux":"bar"},"corge":123},{"qux":{"quux":"baz"},"corge":456}]}`),
		nil,
	},
}

func TestGroup(t *testing.T) {
	ctx := context.TODO()
	capsule := config.NewCapsule()

	for _, test := range groupTests {
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

func benchmarkGroup(b *testing.B, applicator group, test config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = applicator.Apply(ctx, test)
	}
}

func BenchmarkGroup(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range groupTests {
		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkGroup(b, test.proc, capsule)
			},
		)
	}
}
