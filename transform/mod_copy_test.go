package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var _ Transformer = &modCopy{}

var modCopyTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected [][]byte
	err      error
}{
	{
		"object",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key":     "a",
					"set_key": "c",
				},
			},
		},
		[]byte(`{"a":"b"}`),
		[][]byte{
			[]byte(`{"a":"b","c":"b"}`),
		},
		nil,
	},
	{
		"unescape object",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key":     "a",
					"set_key": "a",
				},
			},
		},
		[]byte(`{"a":"{\"b\":\"c\"}"`),
		[][]byte{
			[]byte(`{"a":{"b":"c"}`),
		},
		nil,
	},
	{
		"unescape object",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key":     "a",
					"set_key": "a",
				},
			},
		},
		[]byte(`{"a":"[\"b\",\"c\"]"}`),
		[][]byte{
			[]byte(`{"a":["b","c"]}`),
		},
		nil,
	},
	{
		"from object",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key": "a",
				},
			},
		},
		[]byte(`{"a":"b"}`),
		[][]byte{
			[]byte(`b`),
		},
		nil,
	},
	{
		"from nested object",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key": "a",
				},
			},
		},
		[]byte(`{"a":{"b":"c"}}`),
		[][]byte{
			[]byte(`{"b":"c"}`),
		},
		nil,
	},
	{
		"to object",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"set_key": "a",
				},
			},
		},
		[]byte(`b`),
		[][]byte{
			[]byte(`{"a":"b"}`),
		},
		nil,
	},
	{
		"to object base64",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"set_key": "a",
				},
			},
		},
		[]byte{120, 156, 5, 192, 49, 13, 0, 0, 0, 194, 48, 173, 76, 2, 254, 143, 166, 29, 2, 93, 1, 54},
		[][]byte{
			[]byte(`{"a":"eJwFwDENAAAAwjCtTAL+j6YdAl0BNg=="}`),
		},
		nil,
	},
}

func TestModCopy(t *testing.T) {
	ctx := context.TODO()
	for _, test := range modCopyTests {
		t.Run(test.name, func(t *testing.T) {
			tf, err := newModCopy(ctx, test.cfg)
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

func benchmarkModCopy(b *testing.B, tf *modCopy, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		msg := message.New().SetData(data)
		_, _ = tf.Transform(ctx, msg)
	}
}

func BenchmarkModCopy(b *testing.B) {
	for _, test := range modCopyTests {
		tf, err := newModCopy(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkModCopy(b, tf, test.test)
			},
		)
	}
}
