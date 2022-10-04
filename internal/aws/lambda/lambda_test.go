package lambda

import (
	"bytes"
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/lambda/lambdaiface"
)

type mockedInvoke struct {
	lambdaiface.LambdaAPI
	Resp lambda.InvokeOutput
}

func (m mockedInvoke) InvokeWithContext(ctx aws.Context, input *lambda.InvokeInput, opts ...request.Option) (*lambda.InvokeOutput, error) {
	return &m.Resp, nil
}

func TestPutItem(t *testing.T) {
	tests := []struct {
		resp     lambda.InvokeOutput
		expected []byte
	}{
		{
			resp: lambda.InvokeOutput{
				Payload: []byte("foo"),
			},
			expected: []byte("foo"),
		},
	}

	ctx := context.TODO()

	for _, test := range tests {
		a := API{
			mockedInvoke{Resp: test.resp},
		}

		resp, err := a.Invoke(ctx, "", nil)
		if err != nil {
			t.Fatalf("%d, unexpected error", err)
		}

		if c := bytes.Compare(resp.Payload, test.expected); c != 0 {
			t.Errorf("expected %+v, got %s", resp.Payload, test.expected)
		}
	}
}
