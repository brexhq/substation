package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	mess "github.com/brexhq/substation/message"
)

var _ Transformer = &procBase64{}

var procBase64Tests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected [][]byte
	err      error
}{
	{
		"data decode",
		config.Config{
			Settings: map[string]interface{}{
				"direction": "from",
			},
		},
		[]byte(`YmFy`),
		[][]byte{
			[]byte(`bar`),
		},
		nil,
	},
	{
		"data encode",
		config.Config{
			Settings: map[string]interface{}{
				"direction": "to",
			},
		},
		[]byte(`bar`),
		[][]byte{
			[]byte(`YmFy`),
		},
		nil,
	},
	{
		"JSON decode",
		config.Config{
			Settings: map[string]interface{}{
				"key":       "foo",
				"set_key":   "foo",
				"direction": "from",
			},
		},
		[]byte(`{"foo":"YmFy"}`),
		[][]byte{
			[]byte(`{"foo":"bar"}`),
		},
		nil,
	},
}

func TestprocBase64(t *testing.T) {
	ctx := context.TODO()
	for _, test := range procBase64Tests {
		t.Run(test.name, func(t *testing.T) {
			message, err := mess.New(
				mess.SetData(test.test),
			)
			if err != nil {
				t.Fatal(err)
			}

			proc, err := newProcBase64(ctx, test.cfg)
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

func benchmarkprocBase64(b *testing.B, tf *procBase64, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		message, _ := mess.New(
			mess.SetData(data),
		)

		_, _ = tf.Transform(ctx, message)
	}
}

func BenchmarkprocBase64(b *testing.B) {
	for _, test := range procBase64Tests {
		proc, err := newProcBase64(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkprocBase64(b, proc, test.test)
			},
		)
	}
}
