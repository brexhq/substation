package process

import (
	"bytes"
	"context"
	"errors"
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
	{
		"invalid settings",
		Pipeline{},
		[]byte{},
		[]byte{},
		ProcessorInvalidSettings,
	},
	{
		"invalid array input",
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
		[]byte(`{"foo":["H4sIAMpcy2IA/wXAIQ0AAACAsLbY93csBiFlc4wDAAAA"]}`),
		[]byte{},
		PipelineArrayInput,
	},
}

func TestPipeline(t *testing.T) {
	for _, test := range PipelineTests {
		ctx := context.TODO()
		res, err := test.proc.Byte(ctx, test.test)
		if err != nil && errors.Is(err, test.err) {
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
