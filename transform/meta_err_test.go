package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

var _ Transformer = &metaErr{}

var metaErrTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected [][]byte
}{
	{
		"error_messages string",
		config.Config{
			Settings: map[string]interface{}{
				"transforms": []config.Config{
					{
						Settings: map[string]interface{}{
							"message": "test error",
						},
						Type: "utility_err",
					},
				},
				"error_messages": []string{
					"test error",
				},
			},
		},
		[]byte(`{"a":"b"}`),
		[][]byte{
			[]byte(`{"a":"b"}`),
		},
	},
	{
		"error_messages regex",
		config.Config{
			Settings: map[string]interface{}{
				"transforms": []config.Config{
					{
						Settings: map[string]interface{}{
							"message": "test error",
						},
						Type: "utility_err",
					},
				},
				"error_messages": []string{
					"^test",
				},
			},
		},
		[]byte(`{"a":"b"}`),
		[][]byte{
			[]byte(`{"a":"b"}`),
		},
	},
}

func TestMetaErr(t *testing.T) {
	ctx := context.TODO()
	for _, test := range metaErrTests {
		t.Run(test.name, func(t *testing.T) {
			tf, err := newMetaErr(ctx, test.cfg)
			if err != nil {
				t.Fatal(err)
			}

			msg := message.New().SetData(test.test)
			result, err := tf.Transform(ctx, msg)
			if err != nil {
				t.Fatal(err)
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

func benchmarkMetaErr(b *testing.B, tf *metaErr, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		msg := message.New().SetData(data)
		_, _ = tf.Transform(ctx, msg)
	}
}

func BenchmarkMetaErr(b *testing.B) {
	for _, test := range metaErrTests {
		tf, err := newMetaErr(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkMetaErr(b, tf, test.test)
			},
		)
	}
}

func FuzzTestMetaErr(f *testing.F) {
	testcases := [][]byte{
		[]byte(`{"a":"b"}`),
		[]byte(`{"c":"d"}`),
		[]byte(`{"e":"f"}`),
		[]byte(`{"a":{"b":"c"}}`),
		[]byte(`{"array":[1,2,3]}`),
		[]byte(``),
	}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		ctx := context.TODO()
		msg := message.New().SetData(data)

		// Use a sample configuration for the transformer
		tf, err := newMetaErr(ctx, config.Config{
			Settings: map[string]interface{}{
				"error": "sample error",
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
