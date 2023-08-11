package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	mess "github.com/brexhq/substation/message"
)

var _ Transformer = &metaPipeline{}

var metaPipelineTests = []struct {
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
				"key":     "input",
				"set_key": "output",
				"transforms": []config.Config{
					{
						Type: "proc_base64",
						Settings: map[string]interface{}{
							"direction": "from",
						},
					},
					{
						Type: "proc_gzip",
						Settings: map[string]interface{}{
							"direction": "from",
						},
					},
				},
			},
		},
		[]byte(`{"input":"H4sIAO291GIA/wXAIQ0AAACAsLbY93csBiFlc4wDAAAA"}`),
		[][]byte{
			[]byte(`{"output":"foo"}`),
		},
		nil,
	},
	{
		"data",
		config.Config{
			Settings: map[string]interface{}{
				"transforms": []config.Config{
					{
						Type: "proc_base64",
						Settings: map[string]interface{}{
							"direction": "from",
						},
					},
					{
						Type: "proc_gzip",
						Settings: map[string]interface{}{
							"direction": "from",
						},
					},
				},
			},
		},
		[]byte(`H4sIAO291GIA/wXAIQ0AAACAsLbY93csBiFlc4wDAAAA`),
		[][]byte{
			[]byte(`foo`),
		},
		nil,
	},
}

func TestMetaPipeline(t *testing.T) {
	ctx := context.TODO()
	for _, test := range metaPipelineTests {
		t.Run(test.name, func(t *testing.T) {
			message, err := mess.New(
				mess.SetData(test.test),
			)
			if err != nil {
				t.Fatal(err)
			}

			proc, err := newMetaPipeline(ctx, test.cfg)
			if err != nil {
				t.Fatal(err)
			}

			result, err := proc.Transform(ctx, message)
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
		message, _ := mess.New(
			mess.SetData(data),
		)

		_, _ = tf.Transform(ctx, message)
	}
}

func BenchmarkMetaPipeline(b *testing.B) {
	for _, test := range metaPipelineTests {
		proc, err := newMetaPipeline(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkMetaPipeline(b, proc, test.test)
			},
		)
	}
}
