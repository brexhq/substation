package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var _ Transformer = &modReplace{}

var modReplaceTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected [][]byte
	err      error
}{
	{
		"data replace",
		config.Config{
			Settings: map[string]interface{}{
				"old": "c",
				"new": "b",
			},
		},
		[]byte(`abc`),
		[][]byte{
			[]byte(`abb`),
		},
		nil,
	},
	{
		"data delete",
		config.Config{
			Settings: map[string]interface{}{
				"old": "c",
				"new": "",
			},
		},
		[]byte(`abc`),
		[][]byte{
			[]byte(`ab`),
		},
		nil,
	},
	{
		"object replace",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key":     "a",
					"set_key": "a",
				},
				"old": "c",
				"new": "b",
			},
		},
		[]byte(`{"a":"bc"}`),
		[][]byte{
			[]byte(`{"a":"bb"}`),
		},
		nil,
	},
	{
		"object delete",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key":     "a",
					"set_key": "a",
				},
				"old": "c",
			},
		},
		[]byte(`{"a":"bc"}`),
		[][]byte{
			[]byte(`{"a":"b"}`),
		},
		nil,
	},
}

func TestModReplace(t *testing.T) {
	ctx := context.TODO()
	for _, test := range modReplaceTests {
		t.Run(test.name, func(t *testing.T) {
			tf, err := newModReplace(ctx, test.cfg)
			if err != nil {
				t.Fatal(err)
			}

			msg := message.New().SetData(test.test)
			result, err := tf.Transform(ctx, msg)
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

func benchmarkModReplace(b *testing.B, tf *modReplace, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		msg := message.New().SetData(data)
		_, _ = tf.Transform(ctx, msg)
	}
}

func BenchmarkModReplace(b *testing.B) {
	for _, test := range modReplaceTests {
		tf, err := newModReplace(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkModReplace(b, tf, test.test)
			},
		)
	}
}
