package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	mess "github.com/brexhq/substation/message"
)

var _ Transformer = &procSplit{}

var procSplitTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected [][]byte
	err      error
}{
	{
		"JSON",
		config.Config{
			Settings: map[string]interface{}{
				"key":       "a",
				"set_key":   "a",
				"separator": ".",
			},
		},
		[]byte(`{"a":"b.c.d"}`),
		[][]byte{
			[]byte(`{"a":["b","c","d"]}`),
		},
		nil,
	},
	{
		"data",
		config.Config{
			Settings: map[string]interface{}{
				"separator": `\n`,
			},
		},
		[]byte(`{"a":"b"}\n{"c":"d"}\n{"e":"f"}`),
		[][]byte{
			[]byte(`{"a":"b"}`),
			[]byte(`{"c":"d"}`),
			[]byte(`{"e":"f"}`),
		},
		nil,
	},
}

func TestProcSplit(t *testing.T) {
	ctx := context.TODO()
	for _, test := range procSplitTests {
		t.Run(test.name, func(t *testing.T) {
			message, err := mess.New(
				mess.SetData(test.test),
			)
			if err != nil {
				t.Fatal(err)
			}

			proc, err := newProcSplit(ctx, test.cfg)
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

func benchmarkProcSplit(b *testing.B, tf *procSplit, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		message, _ := mess.New(
			mess.SetData(data),
		)

		_, _ = tf.Transform(ctx, message)
	}
}

func BenchmarkProcSplit(b *testing.B) {
	for _, test := range procSplitTests {
		p, err := newProcSplit(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkProcSplit(b, p, test.test)
			},
		)
	}
}
