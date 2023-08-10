package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	mess "github.com/brexhq/substation/message"
)

var _ Transformer = &procGroup{}

var procGroupTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected [][]byte
	err      error
}{
	{
		"tuples",
		config.Config{
			Settings: map[string]interface{}{
				"key":     "group",
				"set_key": "group",
			},
		},
		[]byte(`{"group":[["foo","bar"],[123,456]]}`),
		[][]byte{
			[]byte(`{"group":[["foo",123],["bar",456]]}`),
		},
		nil,
	},
	{
		"objects",
		config.Config{
			Settings: map[string]interface{}{
				"key":     "group",
				"set_key": "group",
				"keys":    []string{"qux.quux", "corge"},
			},
		},
		[]byte(`{"group":[["foo","bar"],[123,456]]}`),
		[][]byte{
			[]byte(`{"group":[{"qux":{"quux":"foo"},"corge":123},{"qux":{"quux":"bar"},"corge":456}]}`),
		},
		nil,
	},
}

func TestProcGroup(t *testing.T) {
	ctx := context.TODO()
	for _, test := range procGroupTests {
		t.Run(test.name, func(t *testing.T) {
			message, err := mess.New(
				mess.SetData(test.test),
			)
			if err != nil {
				t.Fatal(err)
			}

			proc, err := newProcGroup(ctx, test.cfg)
			if err != nil {
				t.Fatal(err)
			}

			result, err := proc.Transform(ctx, message)
			if err != nil {
				t.Error(err)
			}

			var data [][]byte
			for _, c := range result {
				data = append(data, c.Data())
			}

			if !reflect.DeepEqual(data, test.expected) {
				t.Errorf("expected %s, got %s", test.expected, data)
			}
		})
	}
}

func benchmarkProcGroup(b *testing.B, tform *procGroup, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		message, _ := mess.New(
			mess.SetData(data),
		)

		_, _ = tform.Transform(ctx, message)
	}
}

func BenchmarkProcGroup(b *testing.B) {
	for _, test := range procGroupTests {
		proc, err := newProcGroup(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkProcGroup(b, proc, test.test)
			},
		)
	}
}
