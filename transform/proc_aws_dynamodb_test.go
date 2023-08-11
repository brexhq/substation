package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/brexhq/substation/config"
	ddb "github.com/brexhq/substation/internal/aws/dynamodb"
	mess "github.com/brexhq/substation/message"
)

var _ Transformer = &procAWSDynamoDB{}

type mockedQuery struct {
	dynamodbiface.DynamoDBAPI
	Resp dynamodb.QueryOutput
}

func (m mockedQuery) QueryWithContext(ctx aws.Context, input *dynamodb.QueryInput, opts ...request.Option) (*dynamodb.QueryOutput, error) {
	return &m.Resp, nil
}

var procAWSDynamoDBTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected [][]byte
	err      error
	api      ddb.API
}{
	{
		"JSON",
		config.Config{
			Settings: map[string]interface{}{
				"key":                      "foo",
				"set_key":                  "foo",
				"table":                    "fooer",
				"key_condition_expression": "barre",
			},
		},
		[]byte(`{"foo":{"PK":"bar"}}`),
		[][]byte{
			[]byte(`{"foo":[{"baz":"qux"}]}`),
		},
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

func TestProcAWSDynamoDB(t *testing.T) {
	ctx := context.TODO()
	for _, test := range procAWSDynamoDBTests {
		message, err := mess.New(
			mess.SetData(test.test),
		)
		if err != nil {
			t.Fatal(err)
		}

		proc, err := newProcAWSDynamoDB(ctx, test.cfg)
		if err != nil {
			t.Fatal(err)
		}
		proc.client = test.api

		result, err := proc.Transform(ctx, message)
		if err != nil {
			t.Error(err)
		}

		var data [][]byte
		for _, c := range result {
			data = append(data, c.Data())
		}

		if !reflect.DeepEqual(data, test.expected) {
			t.Errorf("expected %s, got %s", test.expected, data)
		}
	}
}

func benchmarkProcAWSDynamoDB(b *testing.B, tf *procAWSDynamoDB, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		message, _ := mess.New(
			mess.SetData(data),
		)

		_, _ = tf.Transform(ctx, message)
	}
}

func BenchmarkProcAWSDynamoDB(b *testing.B) {
	ctx := context.TODO()
	for _, test := range procAWSDynamoDBTests {
		b.Run(test.name,
			func(b *testing.B) {
				proc, err := newProcAWSDynamoDB(ctx, test.cfg)
				if err != nil {
					b.Fatal(err)
				}
				proc.client = test.api

				benchmarkProcAWSDynamoDB(b, proc, test.test)
			},
		)
	}
}
