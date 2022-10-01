package process

import (
	"bytes"
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var PipelineTests = []struct {
	name     string
	proc     Pipeline
	test     []byte
	expected []byte
	err      error
}{
	{
		"json",
		Pipeline{
			Options: PipelineOptions{
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
			InputKey:  "foo",
			OutputKey: "foo",
		},
		[]byte(`{"foo":"H4sIAKi91GIA/wXAMQ0AAADCMK1MAv6Pph2qjP92AwAAAA=="}`),
		[]byte(`{"foo":"bar"}`),
		nil,
	},
	{
		"data",
		Pipeline{
			Options: PipelineOptions{
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
	cap := config.NewCapsule()

	for _, test := range PipelineTests {
		cap.SetData(test.test)

		result, err := test.proc.Apply(ctx, cap)
		if err != nil {
			t.Log(err)
			t.Fail()
		}

		if !bytes.Equal(result.Data(), test.expected) {
			t.Logf("expected %s, got %s", test.expected, result.Data())
			t.Fail()
		}
	}
}

func benchmarkPipeline(b *testing.B, applicator Pipeline, test config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		applicator.Apply(ctx, test)
	}
}

func BenchmarkPipeline(b *testing.B) {
	cap := config.NewCapsule()
	for _, test := range PipelineTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				cap.SetData(test.test)
				benchmarkPipeline(b, test.proc, cap)
			},
		)
	}
}
