package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var objInsertTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected [][]byte
}{
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
	},
	{
		"array",
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
	},
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
	},
}

func TestObjInsert(t *testing.T) {
	ctx := context.TODO()

	for _, test := range objInsertTests {
		t.Run(test.name, func(t *testing.T) {
			tf, err := newObjInsert(ctx, test.cfg)
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

func benchmarkObjInsert(b *testing.B, tf *objInsert, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		msg := message.New().SetData(data)
		_, _ = tf.Transform(ctx, msg)
	}
}

func BenchmarkObjInsert(b *testing.B) {
	for _, test := range objInsertTests {
		tf, err := newObjInsert(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkObjInsert(b, tf, test.test)
			},
		)
	}
}
