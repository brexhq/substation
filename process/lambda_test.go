package process

import (
	"bytes"
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	golambda "github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/lambda/lambdaiface"
	"github.com/brexhq/substation/config"
	lamb "github.com/brexhq/substation/internal/aws/lambda"
)

type mockedInvoke struct {
	lambdaiface.LambdaAPI
	Resp golambda.InvokeOutput
}

func (m mockedInvoke) InvokeWithContext(ctx aws.Context, input *golambda.InvokeInput, opts ...request.Option) (*golambda.InvokeOutput, error) {
	return &m.Resp, nil
}

var lambdaTests = []struct {
	name     string
	proc     lambda
	test     []byte
	expected []byte
	err      error
	api      lamb.API
}{
	{
		"JSON",
		lambda{
			process: process{
				Key:    "foo",
				SetKey: "foo",
			},
			Options: lambdaOptions{
				FunctionName: "fooer",
			},
		},
		[]byte(`{"foo":{"bar":"baz"}}`),
		[]byte(`{"foo":{"baz":"qux"}}`),
		nil,
		lamb.API{
			Client: mockedInvoke{
				Resp: golambda.InvokeOutput{
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

func benchmarkLambda(b *testing.B, applicator lambda, test config.Capsule) {
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
