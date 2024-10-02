package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

var _ Transformer = &objectToBoolean{}

var objectToBooleanTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected [][]byte
}{
	{
		"float to_bool",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "a",
					"target_key": "a",
				},
			},
		},
		[]byte(`{"a":1.0}`),
		[][]byte{
			[]byte(`{"a":true}`),
		},
	},
	{
		"float to_bool",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "a",
					"target_key": "a",
				},
			},
		},
		[]byte(`{"a":0.0}`),
		[][]byte{
			[]byte(`{"a":false}`),
		},
	},
	{
		"int to_bool",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "a",
					"target_key": "a",
				},
			},
		},
		[]byte(`{"a":1}`),
		[][]byte{
			[]byte(`{"a":true}`),
		},
	},
	{
		"int to_bool",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "a",
					"target_key": "a",
				},
			},
		},
		[]byte(`{"a":0}`),
		[][]byte{
			[]byte(`{"a":false}`),
		},
	},
	{
		"str to_bool",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "a",
					"target_key": "a",
				},
			},
		},
		[]byte(`{"a":"true"}`),
		[][]byte{
			[]byte(`{"a":true}`),
		},
	},
	{
		"str to_bool",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "a",
					"target_key": "a",
				},
			},
		},
		[]byte(`{"a":"false"}`),
		[][]byte{
			[]byte(`{"a":false}`),
		},
	},
}

func TestObjectToBoolean(t *testing.T) {
	ctx := context.TODO()

	for _, test := range objectToBooleanTests {
		t.Run(test.name, func(t *testing.T) {
			tf, err := newObjectToBoolean(ctx, test.cfg)
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

func benchmarkObjectToBoolean(b *testing.B, tf *objectToBoolean, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		msg := message.New().SetData(data)
		_, _ = tf.Transform(ctx, msg)
	}
}

func BenchmarkObjectToBoolean(b *testing.B) {
	for _, test := range objectToBooleanTests {
		tf, err := newObjectToBoolean(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkObjectToBoolean(b, tf, test.test)
			},
		)
	}
}

func FuzzTestObjectToBoolean(f *testing.F) {
	testcases := [][]byte{
		[]byte(`{"a":1.0}`),
		[]byte(`{"a":0.0}`),
		[]byte(`{"a":true}`),
		[]byte(`{"a":false}`),
		[]byte(`{"a":"true"}`),
		[]byte(`{"a":"false"}`),
		[]byte(`{"a":null}`),
		[]byte(`{"a":""}`),
		[]byte(`{"a":"1"}`),
		[]byte(`{"a":"0"}`),
		[]byte(``),
	}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		ctx := context.TODO()
		msg := message.New().SetData(data)

		// Use a sample configuration for the transformer
		tf, err := newObjectToBoolean(ctx, config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "a",
					"target_key": "a",
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
