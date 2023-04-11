package process

import (
	"bytes"
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var (
	_ Applier = procGroup{}
	_ Batcher = procGroup{}
)

var groupTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected []byte
	err      error
}{
	{
		"tuples",
		config.Config{
			Type: "group",
			Settings: map[string]interface{}{
				"key":     "group",
				"set_key": "group",
			},
		},
		[]byte(`{"group":[["foo","bar"],[123,456]]}`),
		[]byte(`{"group":[["foo",123],["bar",456]]}`),
		nil,
	},
	{
		"objects",
		config.Config{
			Type: "group",
			Settings: map[string]interface{}{
				"key":     "group",
				"set_key": "group",
				"options": map[string]interface{}{
					"keys": []string{"qux.quux", "corge"},
				},
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
		t.Run(test.name, func(t *testing.T) {
			capsule.SetData(test.test)

			proc, err := newProcGroup(test.cfg)
			if err != nil {
				t.Fatal(err)
			}

			result, err := proc.Apply(ctx, capsule)
			if err != nil {
				t.Error(err)
			}

			if !bytes.Equal(result.Data(), test.expected) {
				t.Errorf("expected %s, got %s", test.expected, result.Data())
			}
		})
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
		proc, err := newProcGroup(test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkGroup(b, proc, capsule)
			},
		)
	}
}
