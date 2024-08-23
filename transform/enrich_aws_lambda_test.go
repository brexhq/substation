package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/lambda/lambdaiface"
	"github.com/brexhq/substation/v2/config"
	lamb "github.com/brexhq/substation/v2/internal/aws/lambda"
	"github.com/brexhq/substation/v2/message"
)

var _ Transformer = &enrichAWSLambda{}

type enrichAWSLambdaMockedInvoke struct {
	lambdaiface.LambdaAPI
	Resp lambda.InvokeOutput
}

func (m enrichAWSLambdaMockedInvoke) InvokeWithContext(ctx aws.Context, input *lambda.InvokeInput, opts ...request.Option) (*lambda.InvokeOutput, error) {
	return &m.Resp, nil
}

var enrichAWSLambdaTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected [][]byte
	err      error
	api      lamb.API
}{
	{
		"success",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "a",
					"target_key": "a",
				},
				"function_name": "func",
			},
		},
		[]byte(`{"a":{"b":"c"}}`),
		[][]byte{
			[]byte(`{"a":{"d":"e"}}`),
		},
		nil,
		lamb.API{
			Client: enrichAWSLambdaMockedInvoke{
				Resp: lambda.InvokeOutput{
					Payload: []byte(`{"d":"e"}`),
				},
			},
		},
	},
}

func TestEnrichAWSLambda(t *testing.T) {
	ctx := context.TODO()
	for _, test := range enrichAWSLambdaTests {
		tf, err := newEnrichAWSLambda(ctx, test.cfg)
		if err != nil {
			t.Fatal(err)
		}
		tf.client = test.api

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
	}
}

func benchmarkEnrichAWSLambda(b *testing.B, tf *enrichAWSLambda, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		msg := message.New().SetData(data)
		_, _ = tf.Transform(ctx, msg)
	}
}

func BenchmarkEnrichAWSLambda(b *testing.B) {
	ctx := context.TODO()
	for _, test := range enrichAWSLambdaTests {
		b.Run(test.name,
			func(b *testing.B) {
				tf, err := newEnrichAWSLambda(ctx, test.cfg)
				if err != nil {
					b.Fatal(err)
				}
				tf.client = test.api

				benchmarkEnrichAWSLambda(b, tf, test.test)
			},
		)
	}
}
