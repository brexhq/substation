package lambda

import (
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

func TestInvoke(t *testing.T) {
	tests := []struct {
		resp     lambda.InvokeOutput
		expected int64
	}{
		{
			resp: lambda.InvokeOutput{
				StatusCode: aws.Int64(200),
			},
			expected: 200,
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

		if *resp.StatusCode != test.expected {
			t.Errorf("expected %+v, got %d", resp.Payload, test.expected)
		}
	}
}

func TestInvokeAsync(t *testing.T) {
	tests := []struct {
		resp     lambda.InvokeOutput
		expected int64
	}{
		{
			resp: lambda.InvokeOutput{
				StatusCode: aws.Int64(202),
			},
			expected: 202,
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

		if *resp.StatusCode != test.expected {
			t.Errorf("expected %+v, got %d", resp.Payload, test.expected)
		}
	}
}
