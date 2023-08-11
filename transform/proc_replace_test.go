package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	mess "github.com/brexhq/substation/message"
)

var _ Transformer = &procReplace{}

var procReplaceTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected [][]byte
	err      error
}{
	{
		"json",
		config.Config{
			Settings: map[string]interface{}{
				"key":     "replace",
				"set_key": "replace",
				"old":     "r",
				"new":     "z",
			},
		},
		[]byte(`{"replace":"bar"}`),
		[][]byte{
			[]byte(`{"replace":"baz"}`),
		},
		nil,
	},
	{
		"json delete",
		config.Config{
			Settings: map[string]interface{}{
				"key":     "replace",
				"set_key": "replace",
				"old":     "z",
			},
		},
		[]byte(`{"replace":"fizz"}`),
		[][]byte{
			[]byte(`{"replace":"fi"}`),
		},
		nil,
	},
	{
		"data",
		config.Config{
			Settings: map[string]interface{}{
				"old": "r",
				"new": "z",
			},
		},
		[]byte(`bar`),
		[][]byte{
			[]byte(`baz`),
		},
		nil,
	},
	{
		"data delete",
		config.Config{
			Settings: map[string]interface{}{
				"old": "r",
				"new": "",
			},
		},
		[]byte(`bar`),
		[][]byte{
			[]byte(`ba`),
		},
		nil,
	},
}

func TestProcReplace(t *testing.T) {
	ctx := context.TODO()
	for _, test := range procReplaceTests {
		t.Run(test.name, func(t *testing.T) {
			message, err := mess.New(
				mess.SetData(test.test),
			)
			if err != nil {
				t.Fatal(err)
			}

			proc, err := newProcReplace(ctx, test.cfg)
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

func benchmarkProcReplace(b *testing.B, tf *procReplace, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		message, _ := mess.New(
			mess.SetData(data),
		)

		_, _ = tf.Transform(ctx, message)
	}
}

func BenchmarkProcReplace(b *testing.B) {
	for _, test := range procReplaceTests {
		proc, err := newProcReplace(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkProcReplace(b, proc, test.test)
			},
		)
	}
}
