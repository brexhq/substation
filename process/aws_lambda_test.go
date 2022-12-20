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

type mockedInvoke struct {
	lambdaiface.LambdaAPI
	Resp lambda.InvokeOutput
}

func (m mockedInvoke) InvokeWithContext(ctx aws.Context, input *lambda.InvokeInput, opts ...request.Option) (*lambda.InvokeOutput, error) {
	return &m.Resp, nil
}

var lambdaTests = []struct {
	name     string
	proc     _awsLambda
	test     []byte
	expected []byte
	err      error
	api      lamb.API
}{
	{
		"JSON",
		_awsLambda{
			process: process{
				Key:    "foo",
				SetKey: "foo",
			},
			Options: _awsLambdaOptions{
				FunctionName: "fooer",
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

func benchmarkLambda(b *testing.B, applicator _awsLambda, test config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = applicator.Apply(ctx, test)
	}
}

func BenchmarkLambda(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range lambdaTests {
		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkLambda(b, test.proc, capsule)
			},
		)
	}
}
