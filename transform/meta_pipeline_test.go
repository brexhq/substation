package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var _ Transformer = &metaPipeline{}

var metaPipelineTests = []struct {
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
					"key":     "a",
					"set_key": "a",
				},
				"transforms": []config.Config{
					{
						Type:     "format_from_base64",
						Settings: map[string]interface{}{},
					},
					{
						Type:     "format_from_gzip",
						Settings: map[string]interface{}{},
					},
				},
			},
		},
		[]byte(`{"a":"H4sIAO291GIA/wXAIQ0AAACAsLbY93csBiFlc4wDAAAA"}`),
		[][]byte{
			[]byte(`{"a":"foo"}`),
		},
	},
	{
		"data",
		config.Config{
			Settings: map[string]interface{}{
				"transforms": []config.Config{
					{
						Type:     "format_from_base64",
						Settings: map[string]interface{}{},
					},
					{
						Type:     "format_from_gzip",
						Settings: map[string]interface{}{},
					},
				},
			},
		},
		[]byte(`H4sIAO291GIA/wXAIQ0AAACAsLbY93csBiFlc4wDAAAA`),
		[][]byte{
			[]byte(`foo`),
		},
	},
}

func TestMetaPipeline(t *testing.T) {
	ctx := context.TODO()
	for _, test := range metaPipelineTests {
		t.Run(test.name, func(t *testing.T) {
			tf, err := newMetaPipeline(ctx, test.cfg)
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

func benchmarkMetaPipeline(b *testing.B, tf *metaPipeline, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		msg := message.New().SetData(data)
		_, _ = tf.Transform(ctx, msg)
	}
}

func BenchmarkMetaPipeline(b *testing.B) {
	for _, test := range metaPipelineTests {
		tf, err := newMetaPipeline(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkMetaPipeline(b, tf, test.test)
			},
		)
	}
}
