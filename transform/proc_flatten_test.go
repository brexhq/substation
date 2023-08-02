package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	mess "github.com/brexhq/substation/message"
)

var _ Transformer = &procFlatten{}

var procFlattenTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected [][]byte
	err      error
}{
	{
		"json",
		config.Config{
			Type: "proc_flatten",
			Settings: map[string]interface{}{
				"key":     "flatten",
				"set_key": "flatten",
			},
		},
		[]byte(`{"flatten":["foo",["bar"]]}`),
		[][]byte{
			[]byte(`{"flatten":["foo","bar"]}`),
		},
		nil,
	},
	{
		"json deep flatten",
		config.Config{
			Type: "proc_flatten",
			Settings: map[string]interface{}{
				"key":     "flatten",
				"set_key": "flatten",
				"deep":    true,
			},
		},
		[]byte(`{"flatten":[["foo"],[[["bar",[["baz"]]]]]]}`),
		[][]byte{
			[]byte(`{"flatten":["foo","bar","baz"]}`),
		},
		nil,
	},
}

func TestProcFlatten(t *testing.T) {
	ctx := context.TODO()
	for _, test := range procFlattenTests {
		t.Run(test.name, func(t *testing.T) {
			message, err := mess.New(
				mess.SetData(test.test),
			)
			if err != nil {
				t.Fatal(err)
			}

			proc, err := newProcFlatten(ctx, test.cfg)
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

func benchmarkProcFlatten(b *testing.B, tform *procFlatten, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		message, _ := mess.New(
			mess.SetData(data),
		)

		_, _ = tform.Transform(ctx, message)
	}
}

func BenchmarkProcFlatten(b *testing.B) {
	for _, test := range procFlattenTests {
		proc, err := newProcFlatten(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkProcFlatten(b, proc, test.test)
			},
		)
	}
}
