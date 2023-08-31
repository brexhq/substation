package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var _ Transformer = &modHash{}

var modHashTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected [][]byte
	err      error
}{
	{
		"data MD5",
		config.Config{
			Settings: map[string]interface{}{
				"algorithm": "MD5",
			},
		},
		[]byte(`a`),
		[][]byte{
			[]byte(`0cc175b9c0f1b6a831c399e269772661`),
		},
		nil,
	},
	{
		"data SHA256",
		config.Config{
			Settings: map[string]interface{}{
				"algorithm": "SHA256",
			},
		},
		[]byte(`a`),
		[][]byte{
			[]byte(`ca978112ca1bbdcafac231b39a23dc4da786eff8147c4e72b9807785afee48bb`),
		},
		nil,
	},
	{
		"object MD5",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key":     "a",
					"set_key": "a",
				},
				"algorithm": "MD5",
			},
		},
		[]byte(`{"a":"b"}`),
		[][]byte{
			[]byte(`{"a":"92eb5ffee6ae2fec3ad71c777531578f"}`),
		},
		nil,
	},
	{
		"object SHA256",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key":     "a",
					"set_key": "a",
				},
				"algorithm": "SHA256",
			},
		},
		[]byte(`{"a":"b"}`),
		[][]byte{
			[]byte(`{"a":"3e23e8160039594a33894f6564e1b1348bbd7a0088d42c4acb73eeaed59c009d"}`),
		},
		nil,
	},
}

func TestModHash(t *testing.T) {
	ctx := context.TODO()
	for _, test := range modHashTests {
		t.Run(test.name, func(t *testing.T) {
			tf, err := newModHash(ctx, test.cfg)
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

func benchmarkModHash(b *testing.B, tf *modHash, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		msg := message.New().SetData(data)
		_, _ = tf.Transform(ctx, msg)
	}
}

func BenchmarkModHash(b *testing.B) {
	for _, test := range modHashTests {
		tf, err := newModHash(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkModHash(b, tf, test.test)
			},
		)
	}
}
