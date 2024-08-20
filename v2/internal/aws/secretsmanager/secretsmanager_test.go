package secretsmanager

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/secretsmanager/secretsmanageriface"
)

type mockedGetSecret struct {
	secretsmanageriface.SecretsManagerAPI
	Resp secretsmanager.GetSecretValueOutput
}

func (m mockedGetSecret) GetSecretValueWithContext(ctx aws.Context, input *secretsmanager.GetSecretValueInput, opts ...request.Option) (*secretsmanager.GetSecretValueOutput, error) {
	return &m.Resp, nil
}

func TestGetSecret(t *testing.T) {
	tests := []struct {
		resp     secretsmanager.GetSecretValueOutput
		input    string
		expected string
	}{
		{
			resp: secretsmanager.GetSecretValueOutput{
				SecretString: aws.String("foo"),
			},
			input:    "fooer",
			expected: "foo",
		},
	}

	ctx := context.TODO()

	for _, test := range tests {
		a := API{
			mockedGetSecret{Resp: test.resp},
		}

		resp, err := a.GetSecret(ctx, test.input)
		if err != nil {
			t.Fatalf("%d, unexpected error", err)
		}

		if resp != test.expected {
			t.Errorf("expected %+v, got %s", resp, test.expected)
		}
	}
}
