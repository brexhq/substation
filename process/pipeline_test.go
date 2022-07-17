package process

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/brexhq/substation/internal/config"
)

var PipelineTests = []struct {
	name     string
	proc     Pipeline
	err      error
	test     []byte
	expected []byte
}{
	{
		"json",
		Pipeline{
			InputKey:  "pipeline",
			OutputKey: "pipeline",
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
		nil,
		[]byte(`{"pipeline":"H4sIAMpcy2IA/wXAIQ0AAACAsLbY93csBiFlc4wDAAAA"}`),
		[]byte(`{"pipeline":"foo"}`),
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
		nil,
		[]byte(`{"pipeline":"H4sIAMpcy2IA/wXAIQ0AAACAsLbY93csBiFlc4wDAAAA"}`),
		[]byte(`foo`),
	},
	{
		"array",
		Pipeline{
			InputKey:  "pipeline",
			OutputKey: "pipeline",
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
		PipelineArrayInput,
		[]byte(`{"pipeline":["H4sIAMpcy2IA/wXAIQ0AAACAsLbY93csBiFlc4wDAAAA"]}`),
		[]byte{},
	},
}

func TestPipeline(t *testing.T) {
	for _, test := range PipelineTests {
		ctx := context.TODO()
		res, err := test.proc.Byte(ctx, test.test)
		if err != nil && errors.As(err, &test.err) {
			continue
		} else if err != nil {
			t.Log(err)
			t.Fail()
		}

		if c := bytes.Compare(res, test.expected); c != 0 {
			t.Logf("expected %s, got %s", test.expected, res)
			t.Fail()
		}
	}
}

func benchmarkPipelineByte(b *testing.B, byter Pipeline, test []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		byter.Byte(ctx, test)
	}
}

func BenchmarkPipelineByte(b *testing.B) {
	for _, test := range PipelineTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				benchmarkPipelineByte(b, test.proc, test.test)
			},
		)
	}
}
