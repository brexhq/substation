package process

import (
	"bytes"
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var pipelineTests = []struct {
	name     string
	proc     pipeline
	test     []byte
	expected []byte
	err      error
}{
	{
		"json",
		pipeline{
			process: process{
				Key:    "foo",
				SetKey: "foo",
			},
			Options: pipelineOptions{
				Processors: []config.Config{
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
		[]byte(`{"foo":"H4sIAKi91GIA/wXAMQ0AAADCMK1MAv6Pph2qjP92AwAAAA=="}`),
		[]byte(`{"foo":"bar"}`),
		nil,
	},
	{
		"data",
		pipeline{
			Options: pipelineOptions{
				Processors: []config.Config{
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
		[]byte(`H4sIAO291GIA/wXAIQ0AAACAsLbY93csBiFlc4wDAAAA`),
		[]byte(`foo`),
		nil,
	},
}

func TestPipeline(t *testing.T) {
	ctx := context.TODO()
	capsule := config.NewCapsule()

	for _, test := range pipelineTests {
		capsule.SetData(test.test)

		result, err := test.proc.Apply(ctx, capsule)
		if err != nil {
			t.Error(err)
		}

		if !bytes.Equal(result.Data(), test.expected) {
			t.Errorf("expected %s, got %s", test.expected, result.Data())
		}
	}
}

func benchmarkPipeline(b *testing.B, applicator pipeline, test config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = applicator.Apply(ctx, test)
	}
}

func BenchmarkPipeline(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range pipelineTests {
		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkPipeline(b, test.proc, capsule)
			},
		)
	}
}
