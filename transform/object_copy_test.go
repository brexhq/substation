package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

var _ Transformer = &objectCopy{}

var objectCopyTests = []struct {
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
					"source_key": "a",
					"target_key": "c",
				},
			},
		},
		[]byte(`{"a":"b"}`),
		[][]byte{
			[]byte(`{"a":"b","c":"b"}`),
		},
	},
	{
		"unescape object",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "a",
					"target_key": "a",
				},
			},
		},
		[]byte(`{"a":"{\"b\":\"c\"}"`),
		[][]byte{
			[]byte(`{"a":{"b":"c"}`),
		},
	},
	{
		"unescape array",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "a",
					"target_key": "a",
				},
			},
		},
		[]byte(`{"a":"[\"b\",\"c\"]"}`),
		[][]byte{
			[]byte(`{"a":["b","c"]}`),
		},
	},
	{
		"from object",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "a",
				},
			},
		},
		[]byte(`{"a":"b"}`),
		[][]byte{
			[]byte(`b`),
		},
	},
	{
		"from nested object",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "a",
				},
			},
		},
		[]byte(`{"a":{"b":"c"}}`),
		[][]byte{
			[]byte(`{"b":"c"}`),
		},
	},
	{
		"to object",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"target_key": "a",
				},
			},
		},
		[]byte(`b`),
		[][]byte{
			[]byte(`{"a":"b"}`),
		},
	},
	{
		"to nested object",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"target_key": "a.b",
				},
			},
		},
		[]byte(`c`),
		[][]byte{
			[]byte(`{"a":{"b":"c"}}`),
		},
	},
	{
		"to object base64",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"target_key": "a",
				},
			},
		},
		[]byte{120, 156, 5, 192, 49, 13, 0, 0, 0, 194, 48, 173, 76, 2, 254, 143, 166, 29, 2, 93, 1, 54},
		[][]byte{
			[]byte(`{"a":"eJwFwDENAAAAwjCtTAL+j6YdAl0BNg=="}`),
		},
	},
}

func TestObjectCopy(t *testing.T) {
	ctx := context.TODO()
	for _, test := range objectCopyTests {
		t.Run(test.name, func(t *testing.T) {
			tf, err := newObjectCopy(ctx, test.cfg)
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

func benchmarkObjectCopy(b *testing.B, tf *objectCopy, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		msg := message.New().SetData(data)
		_, _ = tf.Transform(ctx, msg)
	}
}

func BenchmarkObjectCopy(b *testing.B) {
	for _, test := range objectCopyTests {
		tf, err := newObjectCopy(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkObjectCopy(b, tf, test.test)
			},
		)
	}
}

func FuzzTestObjectCopy(f *testing.F) {
	testcases := [][]byte{
		[]byte(`{"a":"b"}`),
		[]byte(`{"a":{"b":"c"}}`),
		[]byte(`{"array":[1,2,3]}`),
		[]byte(`{"a":"{\"b\":\"c\"}"}`),
		[]byte(``),
	}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		ctx := context.TODO()
		msg := message.New().SetData(data)

		// Use a sample configuration for the transformer
		tf, err := newObjectCopy(ctx, config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "a",
					"target_key": "c",
				},
			},
		})
		if err != nil {
			return
		}

		_, err = tf.Transform(ctx, msg)
		if err != nil {
			return
		}
	})
}
