package process

import (
	"bytes"
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var groupTests = []struct {
	name     string
	proc     procGroup
	test     []byte
	expected []byte
	err      error
}{
	{
		"tuples",
		procGroup{
			process: process{
				Key:    "group",
				SetKey: "group",
			},
		},
		[]byte(`{"group":[["foo","bar"],[123,456]]}`),
		[]byte(`{"group":[["foo",123],["bar",456]]}`),
		nil,
	},
	{
		"objects",
		procGroup{
			process: process{
				Key:    "group",
				SetKey: "group",
			},
			Options: procGroupOptions{
				Keys: []string{"qux.quux", "corge"},
			},
		},
		[]byte(`{"group":[["foo","bar"],[123,456]]}`),
		[]byte(`{"group":[{"qux":{"quux":"foo"},"corge":123},{"qux":{"quux":"bar"},"corge":456}]}`),
		nil,
	},
}

func TestGroup(t *testing.T) {
	ctx := context.TODO()
	capsule := config.NewCapsule()

	for _, test := range groupTests {
		var _ Applier = test.proc
		var _ Batcher = test.proc

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

func benchmarkGroup(b *testing.B, applier procGroup, test config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = applier.Apply(ctx, test)
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
