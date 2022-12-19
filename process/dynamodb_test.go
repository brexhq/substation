package process

import (
	"bytes"
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	sddb "github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/brexhq/substation/config"
	ddb "github.com/brexhq/substation/internal/aws/dynamodb"
)

type mockedQuery struct {
	dynamodbiface.DynamoDBAPI
	Resp sddb.QueryOutput
}

func (m mockedQuery) QueryWithContext(ctx aws.Context, input *sddb.QueryInput, opts ...request.Option) (*sddb.QueryOutput, error) {
	return &m.Resp, nil
}

var dynamodbTests = []struct {
	name     string
	proc     dynamodb
	test     []byte
	expected []byte
	err      error
	api      ddb.API
}{
	{
		"JSON",
		dynamodb{
			process: process{
				Key:    "foo",
				SetKey: "foo",
			},
			Options: dynamodbOptions{
				Table:                  "fooer",
				KeyConditionExpression: "barre",
			},
		},
		[]byte(`{"foo":{"PK":"bar"}}`),
		[]byte(`{"foo":[{"baz":"qux"}]}`),
		nil,
		ddb.API{
			Client: mockedQuery{
				Resp: sddb.QueryOutput{
					Items: []map[string]*sddb.AttributeValue{
						{
							"baz": {
								S: aws.String("qux"),
							},
						},
					},
				},
			},
		},
	},
}

func TestDynamoDB(t *testing.T) {
	ctx := context.TODO()
	capsule := config.NewCapsule()

	for _, test := range dynamodbTests {
		dynamodbAPI = test.api
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

func benchmarkDynamoDB(b *testing.B, applicator dynamodb, test config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = applicator.Apply(ctx, test)
	}
}

func BenchmarkDynamoDB(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range dynamodbTests {
		b.Run(test.name,
			func(b *testing.B) {
				dynamodbAPI = test.api
				capsule.SetData(test.test)
				benchmarkDynamoDB(b, test.proc, capsule)
			},
		)
	}
}
