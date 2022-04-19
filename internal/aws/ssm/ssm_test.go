package ssm

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
)

type mockedGetParameter struct {
	ssmiface.SSMAPI
	Resp ssm.GetParameterOutput
}

func (m mockedGetParameter) GetParameterWithContext(ctx aws.Context, input *ssm.GetParameterInput, opts ...request.Option) (*ssm.GetParameterOutput, error) {
	return &m.Resp, nil
}

func TestGetParameter(t *testing.T) {
	var tests = []struct {
		resp     ssm.GetParameterOutput
		input    string
		expected string
	}{
		{
			resp: ssm.GetParameterOutput{
				Parameter: &ssm.Parameter{
					Value: aws.String("foo"),
				},
			},
			input:    "fooer",
			expected: "foo",
		},
	}

	ctx := context.TODO()

	for _, test := range tests {
		a := API{
			mockedGetParameter{Resp: test.resp},
		}

		resp, err := a.GetParameter(ctx, test.input)
		if err != nil {
			t.Fatalf("%d, unexpected error", err)
		}

		if resp != test.expected {
			t.Logf("expected %+v, got %s", resp, test.expected)
			t.Fail()
		}
	}
}
