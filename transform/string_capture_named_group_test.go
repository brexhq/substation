package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var _ Transformer = &stringCaptureNamedGroup{}

var stringCaptureNamedGroupTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected [][]byte
}{
	// data tests
	{
		"data",
		config.Config{
			Settings: map[string]interface{}{
				"pattern": "(?P<b>[a-zA-Z]+) (?P<d>[a-zA-Z]+)",
			},
		},
		[]byte(`c e`),
		[][]byte{
			[]byte(`{"b":"c","d":"e"}`),
		},
	},
	// object tests
	{
		"object",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key":     "a",
					"set_key": "a",
				},
				"pattern": "(?P<b>[a-zA-Z]+) (?P<d>[a-zA-Z]+)",
			},
		},
		[]byte(`{"a":"c e"}`),
		[][]byte{
			[]byte(`{"a":{"b":"c","d":"e"}}`),
		},
	},
}

func TestStringCaptureNamedGroup(t *testing.T) {
	ctx := context.TODO()
	for _, test := range stringCaptureNamedGroupTests {
		t.Run(test.name, func(t *testing.T) {
			tf, err := newStringCaptureNamedGroup(ctx, test.cfg)
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

func benchmarkStringCaptureNamedGroup(b *testing.B, tf *stringCaptureNamedGroup, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		msg := message.New().SetData(data)
		_, _ = tf.Transform(ctx, msg)
	}
}

func BenchmarkStringCaptureNamedGroup(b *testing.B) {
	for _, test := range stringCaptureNamedGroupTests {
		tf, err := newStringCaptureNamedGroup(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkStringCaptureNamedGroup(b, tf, test.test)
			},
		)
	}
}
