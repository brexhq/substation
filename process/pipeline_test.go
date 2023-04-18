package process

import (
	"bytes"
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var (
	_ Applier = procPipeline{}
	_ Batcher = procPipeline{}
)

var pipelineTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected []byte
	err      error
}{
	{
		"json",
		config.Config{
			Type: "pipeline",
			Settings: map[string]interface{}{
				"key":     "pipeline",
				"set_key": "pipeline",
				"options": map[string]interface{}{
					"processors": []config.Config{
						{
							Type: "base64",
							Settings: map[string]interface{}{
								"options": map[string]interface{}{
									"direction": "from",
								},
							},
						},
						{
							Type: "gzip",
							Settings: map[string]interface{}{
								"options": map[string]interface{}{
									"direction": "from",
								},
							},
						},
					},
				},
			},
		},
		[]byte(`{"pipeline":"H4sIAKi91GIA/wXAMQ0AAADCMK1MAv6Pph2qjP92AwAAAA=="}`),
		[]byte(`{"pipeline":"bar"}`),
		nil,
	},
	{
		"data",
		config.Config{
			Type: "pipeline",
			Settings: map[string]interface{}{
				"options": map[string]interface{}{
					"processors": []config.Config{
						{
							Type: "base64",
							Settings: map[string]interface{}{
								"options": map[string]interface{}{
									"direction": "from",
								},
							},
						},
						{
							Type: "gzip",
							Settings: map[string]interface{}{
								"options": map[string]interface{}{
									"direction": "from",
								},
							},
						},
					},
				},
			},
		},
		[]byte(`H4sIAO291GIA/wXAIQ0AAACAsLbY93csBiFlc4wDAAAA`),
		[]byte(`foo`),
		nil,
	},
}

func TestPipeline(t *testing.T) {
	ctx := context.TODO()
	capsule := config.NewCapsule()

	for _, test := range pipelineTests {
		t.Run(test.name, func(t *testing.T) {
			capsule.SetData(test.test)

			proc, err := newProcPipeline(ctx, test.cfg)
			if err != nil {
				t.Fatal(err)
			}

			result, err := proc.Apply(ctx, capsule)
			if err != nil {
				t.Error(err)
			}

			if !bytes.Equal(result.Data(), test.expected) {
				t.Errorf("expected %s, got %s", test.expected, result.Data())
			}
		})
	}
}

func benchmarkPipeline(b *testing.B, applier procPipeline, test config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = applier.Apply(ctx, test)
	}
}

func BenchmarkPipeline(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range pipelineTests {
		proc, err := newProcPipeline(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkPipeline(b, proc, capsule)
			},
		)
	}
}
