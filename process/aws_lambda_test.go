package process

import (
	"bytes"
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/lambda/lambdaiface"
	"github.com/brexhq/substation/config"
	lamb "github.com/brexhq/substation/internal/aws/lambda"
)

var (
	_ Applier  = procAWSLambda{}
	_ Batcher  = procAWSLambda{}
	_ Streamer = procAWSLambda{}
)

type mockedInvoke struct {
	lambdaiface.LambdaAPI
	Resp lambda.InvokeOutput
}

func (m mockedInvoke) InvokeWithContext(ctx aws.Context, input *lambda.InvokeInput, opts ...request.Option) (*lambda.InvokeOutput, error) {
	return &m.Resp, nil
}

var lambdaTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected []byte
	err      error
	api      lamb.API
}{
	{
		"JSON",
		config.Config{
			Type: "aws_lambda",
			Settings: map[string]interface{}{
				"key":     "foo",
				"set_key": "foo",
				"options": map[string]interface{}{
					"function_name": "fooer",
				},
			},
		},
		[]byte(`{"foo":{"bar":"baz"}}`),
		[]byte(`{"foo":{"baz":"qux"}}`),
		nil,
		lamb.API{
			Client: mockedInvoke{
				Resp: lambda.InvokeOutput{
					Payload: []byte(`{"baz":"qux"}`),
				},
			},
		},
	},
}

func TestLambda(t *testing.T) {
	ctx := context.TODO()
	capsule := config.NewCapsule()

	for _, test := range lambdaTests {
		lambdaAPI = test.api
		proc, err := newProcAWSLambda(ctx, test.cfg)
		if err != nil {
			t.Fatal(err)
		}

		capsule.SetData(test.test)
		result, err := proc.Apply(ctx, capsule)
		if err != nil {
			t.Error(err)
		}

		if !bytes.Equal(result.Data(), test.expected) {
			t.Errorf("expected %s, got %s", test.expected, result.Data())
		}
	}
}

func benchmarkLambda(b *testing.B, applier procAWSLambda, test config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = applier.Apply(ctx, test)
	}
}

func BenchmarkLambda(b *testing.B) {
	ctx := context.TODO()
	capsule := config.NewCapsule()

	for _, test := range lambdaTests {
		b.Run(test.name,
			func(b *testing.B) {
				proc, err := newProcAWSLambda(ctx, test.cfg)
				if err != nil {
					b.Fatal(err)
				}

				capsule.SetData(test.test)
				benchmarkLambda(b, proc, capsule)
			},
		)
	}
}
