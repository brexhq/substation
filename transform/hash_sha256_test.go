package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var _ Transformer = &hashSHA256{}

var hashSHA256Tests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected [][]byte
}{
	{
		"data",
		config.Config{},
		[]byte(`a`),
		[][]byte{
			[]byte(`ca978112ca1bbdcafac231b39a23dc4da786eff8147c4e72b9807785afee48bb`),
		},
	},
	{
		"object",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key":     "a",
					"set_key": "a",
				},
			},
		},
		[]byte(`{"a":"b"}`),
		[][]byte{
			[]byte(`{"a":"3e23e8160039594a33894f6564e1b1348bbd7a0088d42c4acb73eeaed59c009d"}`),
		},
	},
}

func TestHashSHA256(t *testing.T) {
	ctx := context.TODO()
	for _, test := range hashSHA256Tests {
		t.Run(test.name, func(t *testing.T) {
			tf, err := newHashSHA256(ctx, test.cfg)
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

func benchmarkHashSHA256(b *testing.B, tf *hashSHA256, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		msg := message.New().SetData(data)
		_, _ = tf.Transform(ctx, msg)
	}
}

func BenchmarkHashSHA256(b *testing.B) {
	for _, test := range hashSHA256Tests {
		tf, err := newHashSHA256(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkHashSHA256(b, tf, test.test)
			},
		)
	}
}
