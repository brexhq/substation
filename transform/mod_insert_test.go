package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var _ Transformer = &modInsert{}

var modInsertTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected [][]byte
	err      error
}{
	{
		"string",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"set_key": "a",
				},
				"value": "b",
			},
		},
		[]byte{},
		[][]byte{
			[]byte(`{"a":"b"}`),
		},
		nil,
	},
	{
		"int",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"set_key": "a",
				}, "value": 1,
			},
		},
		[]byte(`{"a":"b"}`),
		[][]byte{
			[]byte(`{"a":1}`),
		},
		nil,
	},
	{
		"string array",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"set_key": "a",
				}, "value": []string{"b", "c"},
			},
		},
		[]byte(`{"a":"b"}`),
		[][]byte{
			[]byte(`{"a":["b","c"]}`),
		},
		nil,
	},
	{
		"map",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"set_key": "a",
				}, "value": map[string]string{
					"b": "c",
				},
			},
		},
		[]byte(`{"a":"b"}`),
		[][]byte{
			[]byte(`{"a":{"b":"c"}}`),
		},
		nil,
	},
	{
		"object",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"set_key": "a",
				}, "value": `{"b":"c"}`,
			},
		},
		[]byte(`{"a":"b"}`),
		[][]byte{
			[]byte(`{"a":{"b":"c"}}`),
		},
		nil,
	},
	{
		"bytes",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"set_key": "a",
				},
				"value": []byte{120, 156, 5, 192, 49, 13, 0, 0, 0, 194, 48, 173, 76, 2, 254, 143, 166, 29, 2, 93, 1, 54},
			},
		},
		[]byte(`{"a":"b"}`),
		[][]byte{
			[]byte(`{"a":"eJwFwDENAAAAwjCtTAL+j6YdAl0BNg=="}`),
		},
		nil,
	},
}

func TestModInsert(t *testing.T) {
	ctx := context.TODO()

	for _, test := range modInsertTests {
		t.Run(test.name, func(t *testing.T) {
			tf, err := newModInsert(ctx, test.cfg)
			if err != nil {
				t.Fatal(err)
			}

			msg := message.New().SetData(test.test)
			result, err := tf.Transform(ctx, msg)
			if err != nil {
				t.Error(err)
			}

			var r [][]byte
			for _, c := range result {
				r = append(r, c.Data())
			}

			if !reflect.DeepEqual(r, test.expected) {
				t.Errorf("expected %s, got %s", test.expected, r)
			}
		})
	}
}

func benchmarkModInsert(b *testing.B, tf *modInsert, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		msg := message.New().SetData(data)
		_, _ = tf.Transform(ctx, msg)
	}
}

func BenchmarkModInsert(b *testing.B) {
	for _, test := range modInsertTests {
		tf, err := newModInsert(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkModInsert(b, tf, test.test)
			},
		)
	}
}