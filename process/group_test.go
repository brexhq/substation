package process

import (
	"bytes"
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var groupTests = []struct {
	name     string
	proc     Group
	test     []byte
	expected []byte
	err      error
}{
	{
		"tuples",
		Group{
			InputKey:  "group",
			OutputKey: "group",
		},
		[]byte(`{"group":[["foo","bar"],[123,456]]}`),
		[]byte(`{"group":[["foo",123],["bar",456]]}`),
		nil,
	},
	{
		"objects",
		Group{
			Options: GroupOptions{
				Keys: []string{"name.test", "size"},
			},
			InputKey:  "group",
			OutputKey: "group",
		},
		[]byte(`{"group":[["foo","bar"],[123,456]]}`),
		[]byte(`{"group":[{"name":{"test":"foo"},"size":123},{"name":{"test":"bar"},"size":456}]}`),
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

func benchmarkGroup(b *testing.B, applicator Group, test config.Capsule) {
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
