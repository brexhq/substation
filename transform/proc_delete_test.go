package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	mess "github.com/brexhq/substation/message"
)

var _ Transformer = &procDelete{}

var procDeleteTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected [][]byte
	err      error
}{
	{
		"string",
		config.Config{
			Type: "delete",
			Settings: map[string]interface{}{
				"key": "baz",
			},
		},
		[]byte(`{"foo":"bar","baz":"qux"}`),
		[][]byte{
			[]byte(`{"foo":"bar"}`),
		},
		nil,
	},
	{
		"JSON",
		config.Config{
			Type: "delete",
			Settings: map[string]interface{}{
				"key": "baz",
			},
		},
		[]byte(`{"foo":"bar","baz":{"qux":"quux"}}`),
		[][]byte{
			[]byte(`{"foo":"bar"}`),
		},
		nil,
	},
}

func TestProcDelete(t *testing.T) {
	ctx := context.TODO()
	for _, test := range procDeleteTests {
		t.Run(test.name, func(t *testing.T) {
			message, err := mess.New(
				mess.SetData(test.test),
			)
			if err != nil {
				t.Fatal(err)
			}

			proc, err := newProcDelete(ctx, test.cfg)
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

func benchmarkProcDelete(b *testing.B, tform *procDelete, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		message, _ := mess.New(
			mess.SetData(data),
		)

		_, _ = tform.Transform(ctx, message)
	}
}

func BenchmarkProcDelete(b *testing.B) {
	for _, test := range procDeleteTests {
		proc, err := newProcDelete(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkProcDelete(b, proc, test.test)
			},
		)
	}
}
