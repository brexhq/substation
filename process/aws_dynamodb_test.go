package process

import (
	"bytes"
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/brexhq/substation/config"
	ddb "github.com/brexhq/substation/internal/aws/dynamodb"
)

type mockedQuery struct {
	dynamodbiface.DynamoDBAPI
	Resp dynamodb.QueryOutput
}

func (m mockedQuery) QueryWithContext(ctx aws.Context, input *dynamodb.QueryInput, opts ...request.Option) (*dynamodb.QueryOutput, error) {
	return &m.Resp, nil
}

var dynamodbTests = []struct {
	name     string
	proc     procAWSDynamoDB
	test     []byte
	expected []byte
	err      error
	api      ddb.API
}{
	{
		"JSON",
		procAWSDynamoDB{
			process: process{
				Key:    "foo",
				SetKey: "foo",
			},
			Options: procAWSDynamoDBOptions{
				Table:                  "fooer",
				KeyConditionExpression: "barre",
			},
		},
		[]byte(`{"foo":{"PK":"bar"}}`),
		[]byte(`{"foo":[{"baz":"qux"}]}`),
		nil,
		ddb.API{
			Client: mockedQuery{
				Resp: dynamodb.QueryOutput{
					Items: []map[string]*dynamodb.AttributeValue{
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
		var _ Applier = test.proc
		var _ Batcher = test.proc

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

func benchmarkDynamoDB(b *testing.B, applier procAWSDynamoDB, test config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = applier.Apply(ctx, test)
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
