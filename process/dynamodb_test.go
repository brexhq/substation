package process

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
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
	proc     DynamoDB
	err      error
	test     []byte
	expected []byte
	api      ddb.API
}{
	{
		"ddb",
		DynamoDB{
			InputKey:  "ddb",
			OutputKey: "ddb",
			Options: DynamoDBOptions{
				Table: "test",
			},
		},
		nil,
		[]byte(`{"ddb":{"PK":"foo"}}`),
		[]byte(`{"ddb":[{"foo":"bar"}]}`),
		ddb.API{
			Client: mockedQuery{
				Resp: dynamodb.QueryOutput{
					Items: []map[string]*dynamodb.AttributeValue{
						{
							"foo": {
								S: aws.String("bar"),
							},
						},
					},
				},
			},
		},
	},
	{
		"invalid settings",
		DynamoDB{},
		ProcessorInvalidSettings,
		[]byte(`{"ddb":{"PK":"foo"}}`),
		[]byte{},
		ddb.API{
			Client: mockedQuery{
				Resp: dynamodb.QueryOutput{
					Items: []map[string]*dynamodb.AttributeValue{
						{
							"foo": {
								S: aws.String("bar"),
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
	for _, test := range dynamodbTests {
		dynamodbAPI = test.api
		res, err := test.proc.Byte(ctx, test.test)
		if err != nil && errors.As(err, &test.err) {
			continue
		} else if err != nil {
			t.Log(err)
			t.Fail()
		}

		if c := bytes.Compare(res, test.expected); c != 0 {
			t.Logf("expected %s, got %s", test.expected, res)
			t.Fail()
		}
	}
}

func benchmarkDynamoDBByte(b *testing.B, byter DynamoDB, test []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		byter.Byte(ctx, test)
	}
}

func BenchmarkDynamoDBByte(b *testing.B) {
	for _, test := range dynamodbTests {
		dynamodbAPI = test.api
		b.Run(string(test.name),
			func(b *testing.B) {
				benchmarkDynamoDBByte(b, test.proc, test.test)
			},
		)
	}
}
