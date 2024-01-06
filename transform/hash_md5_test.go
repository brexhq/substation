package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var _ Transformer = &hashMD5{}

var hashMD5Tests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected [][]byte
}{
	{
		"data",
		config.Config{
			Settings: map[string]interface{}{
				"algorithm": "MD5",
			},
		},
		[]byte(`a`),
		[][]byte{
			[]byte(`0cc175b9c0f1b6a831c399e269772661`),
		},
	},
	{
		"object",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "a",
					"target_key": "a",
				},
				"algorithm": "MD5",
			},
		},
		[]byte(`{"a":"b"}`),
		[][]byte{
			[]byte(`{"a":"92eb5ffee6ae2fec3ad71c777531578f"}`),
		},
	},
}

func TestHashMD5(t *testing.T) {
	ctx := context.TODO()
	for _, test := range hashMD5Tests {
		t.Run(test.name, func(t *testing.T) {
			tf, err := newHashMD5(ctx, test.cfg)
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

func benchmarkHashMD5(b *testing.B, tf *hashMD5, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		msg := message.New().SetData(data)
		_, _ = tf.Transform(ctx, msg)
	}
}

func BenchmarkHashMD5(b *testing.B) {
	for _, test := range hashMD5Tests {
		tf, err := newHashMD5(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkHashMD5(b, tf, test.test)
			},
		)
	}
}
