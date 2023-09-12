package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var stringPatternNamedGroupTests = []struct {
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

func TestStringPatternNamedGroup(t *testing.T) {
	ctx := context.TODO()
	for _, test := range stringPatternNamedGroupTests {
		t.Run(test.name, func(t *testing.T) {
			tf, err := newStringPatternNamedGroup(ctx, test.cfg)
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

func benchmarkStringPatternNamedGroup(b *testing.B, tf *stringPatternNamedGroup, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		msg := message.New().SetData(data)
		_, _ = tf.Transform(ctx, msg)
	}
}

func BenchmarkStringPatternNamedGroup(b *testing.B) {
	for _, test := range stringPatternNamedGroupTests {
		tf, err := newStringPatternNamedGroup(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkStringPatternNamedGroup(b, tf, test.test)
			},
		)
	}
}
