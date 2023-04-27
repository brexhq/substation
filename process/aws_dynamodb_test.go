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

var (
	_ Applier  = procAWSDynamoDB{}
	_ Batcher  = procAWSDynamoDB{}
	_ Streamer = procAWSDynamoDB{}
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
	cfg      config.Config
	test     []byte
	expected []byte
	err      error
	api      ddb.API
}{
	{
		"JSON",
		config.Config{
			Type: "aws_dynamodb",
			Settings: map[string]interface{}{
				"key":     "foo",
				"set_key": "foo",
				"options": map[string]interface{}{
					"table":                    "fooer",
					"key_condition_expression": "barre",
				},
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
		dynamodbAPI = test.api

		proc, err := newProcAWSDynamoDB(ctx, test.cfg)
		if err != nil {
			t.Fatal(err)
		}

		capsule.SetData(test.test)
		result, err := proc.Apply(ctx, capsule)
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
	ctx := context.TODO()
	capsule := config.NewCapsule()

	for _, test := range dynamodbTests {
		b.Run(test.name,
			func(b *testing.B) {
				dynamodbAPI = test.api
				proc, err := newProcAWSDynamoDB(ctx, test.cfg)
				if err != nil {
					b.Fatal(err)
				}

				capsule.SetData(test.test)
				benchmarkDynamoDB(b, proc, capsule)
			},
		)
	}
}
