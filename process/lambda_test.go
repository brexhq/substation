package process

import (
	"bytes"
	"context"
	"errors"
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
	proc     Lambda
	test     []byte
	expected []byte
	err      error
	api      lamb.API
}{
	{
		"JSON",
		Lambda{
			Options: LambdaOptions{
				Function: "fooer",
			},
			InputKey:  "foo",
			OutputKey: "foo",
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
	{
		"invalid settings",
		Lambda{},
		[]byte{},
		[]byte{},
		ProcessorInvalidSettings,
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
	for _, test := range lambdaTests {

		lambdaAPI = test.api
		cap := config.NewCapsule()
		cap.SetData(test.test)

		res, err := test.proc.Apply(ctx, cap)
		if err != nil && errors.Is(err, test.err) {
			continue
		} else if err != nil {
			t.Log(err)
			t.Fail()
		}

		if c := bytes.Compare(res.GetData(), test.expected); c != 0 {
			t.Logf("expected %s, got %s", test.expected, res.GetData())
			t.Fail()
		}
	}
}

func benchmarkLambdaCapByte(b *testing.B, applicator Lambda, test config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		applicator.Apply(ctx, test)
	}
}

func BenchmarkLambdaCapByte(b *testing.B) {
	for _, test := range lambdaTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				cap := config.NewCapsule()
				cap.SetData(test.test)
				benchmarkLambdaCapByte(b, test.proc, cap)
			},
		)
	}
}
