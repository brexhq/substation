package process

import (
	"bytes"
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/lambda/lambdaiface"
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
	api      lamb.API
}{
	{
		"json",
		Lambda{
			Input: LambdaInput{
				Payload: []struct {
					Key        string "json:\"key\""
					PayloadKey string "json:\"payload_key\""
				}{{
					Key:        "foo",
					PayloadKey: "foo",
				},
				},
			},
			Options: LambdaOptions{
				Function: "test",
			},
			Output: Output{
				Key: "lambda",
			},
		},
		[]byte(`{"foo":"bar"}`),
		[]byte(`{"foo":"bar","lambda":{"baz":"qux"}}`),
		lamb.API{
			mockedInvoke{
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
		res, err := test.proc.Byte(ctx, test.test)
		if err != nil {
			t.Logf("%v", err)
			t.Fail()
		}

		if c := bytes.Compare(res, test.expected); c != 0 {
			t.Logf("expected %s, got %s", test.expected, res)
			t.Fail()
		}
	}
}

func benchmarkLambdaByte(b *testing.B, byter Lambda, test []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		byter.Byte(ctx, test)
	}
}

func BenchmarkLambdaByte(b *testing.B) {
	for _, test := range lambdaTests {
		lambdaAPI = test.api
		b.Run(string(test.name),
			func(b *testing.B) {
				benchmarkLambdaByte(b, test.proc, test.test)
			},
		)
	}
}
