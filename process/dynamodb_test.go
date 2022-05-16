package process

import (
	"bytes"
	"context"
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
	test     []byte
	expected []byte
	api      ddb.API
}{
	{
		"json",
		DynamoDB{
			Input: DynamoDBInput{
				PartitionKey: "pk",
			},
			Options: DynamoDBOptions{
				Table: "test",
			},
			Output: Output{
				Key: "ddb",
			},
		},
		[]byte(`{"pk":"foo"}`),
		[]byte(`{"pk":"foo","ddb":[{"foo":"bar"}]}`),
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
	for _, test := range dynamodbTests {
		ctx := context.TODO()
		dynamodbAPI = test.api
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
