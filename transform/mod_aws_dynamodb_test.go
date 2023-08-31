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
	"github.com/brexhq/substation/message"
)

var _ Transformer = &modAWSDynamoDB{}

type mockedQuery struct {
	dynamodbiface.DynamoDBAPI
	Resp dynamodb.QueryOutput
}

func (m mockedQuery) QueryWithContext(ctx aws.Context, input *dynamodb.QueryInput, opts ...request.Option) (*dynamodb.QueryOutput, error) {
	return &m.Resp, nil
}

var modAWSDynamoDBTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected [][]byte
	err      error
	api      ddb.API
}{
	{
		"success",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"set_key": "a",
				},
				"table":                    "tab",
				"partition_key":            "PK",
				"key_condition_expression": "kce",
			},
		},
		[]byte(`{"PK":"b"}`),
		[][]byte{
			[]byte(`{"PK":"b","a":[{"b":"c"}]}`),
		},
		nil,
		ddb.API{
			Client: mockedQuery{
				Resp: dynamodb.QueryOutput{
					Items: []map[string]*dynamodb.AttributeValue{
						{
							"b": {
								S: aws.String("c"),
							},
						},
					},
				},
			},
		},
	},
}

func TestModAWSDynamoDB(t *testing.T) {
	ctx := context.TODO()
	for _, test := range modAWSDynamoDBTests {
		tf, err := newModAWSDynamoDB(ctx, test.cfg)
		if err != nil {
			t.Fatal(err)
		}
		tf.client = test.api

		msg := message.New().SetData(test.test)
		result, err := tf.Transform(ctx, msg)
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

func benchmarkModAWSDynamoDB(b *testing.B, tf *modAWSDynamoDB, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		msg := message.New().SetData(data)
		_, _ = tf.Transform(ctx, msg)
	}
}

func BenchmarkModAWSDynamoDB(b *testing.B) {
	ctx := context.TODO()
	for _, test := range modAWSDynamoDBTests {
		b.Run(test.name,
			func(b *testing.B) {
				tf, err := newModAWSDynamoDB(ctx, test.cfg)
				if err != nil {
					b.Fatal(err)
				}
				tf.client = test.api

				benchmarkModAWSDynamoDB(b, tf, test.test)
			},
		)
	}
}
