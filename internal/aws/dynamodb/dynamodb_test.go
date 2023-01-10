package dynamodb

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

type mockedGetItem struct {
	dynamodbiface.DynamoDBAPI
	Resp dynamodb.GetItemOutput
}

func (m mockedGetItem) GetItemWithContext(ctx aws.Context, input *dynamodb.GetItemInput, opts ...request.Option) (*dynamodb.GetItemOutput, error) {
	return &m.Resp, nil
}

func TestGetItem(t *testing.T) {
	tests := []struct {
		resp     dynamodb.GetItemOutput
		expected string
	}{
		{
			resp: dynamodb.GetItemOutput{
				Item: map[string]*dynamodb.AttributeValue{
					"foo": {
						S: aws.String("bar"),
					},
				},
				ConsumedCapacity: &dynamodb.ConsumedCapacity{},
			},
			expected: "bar",
		},
	}

	ctx := context.TODO()

	for _, test := range tests {
		a := API{
			mockedGetItem{Resp: test.resp},
		}

		m := make(map[string]interface{})
		resp, err := a.GetItem(ctx, "", m)
		if err != nil {
			t.Fatalf("%d, unexpected error", err)
		}

		var item map[string]interface{}
		err = dynamodbattribute.UnmarshalMap(resp.Item, &item)
		if err != nil {
			t.Fatalf("%v, unexpected error", err)
		}

		if item["foo"] != test.expected {
			t.Errorf("expected %+v, got %s", item["foo"], test.expected)
		}
	}
}

type mockedPutItem struct {
	dynamodbiface.DynamoDBAPI
	Resp dynamodb.PutItemOutput
}

func (m mockedPutItem) PutItemWithContext(ctx aws.Context, input *dynamodb.PutItemInput, opts ...request.Option) (*dynamodb.PutItemOutput, error) {
	return &m.Resp, nil
}

func TestPutItem(t *testing.T) {
	tests := []struct {
		resp     dynamodb.PutItemOutput
		expected string
	}{
		{
			resp: dynamodb.PutItemOutput{
				Attributes: map[string]*dynamodb.AttributeValue{
					"foo": {
						S: aws.String("bar"),
					},
				},
				ConsumedCapacity:      &dynamodb.ConsumedCapacity{},
				ItemCollectionMetrics: &dynamodb.ItemCollectionMetrics{},
			},
			expected: "bar",
		},
	}

	ctx := context.TODO()

	for _, test := range tests {
		a := API{
			mockedPutItem{Resp: test.resp},
		}

		resp, err := a.PutItem(ctx, "", map[string]*dynamodb.AttributeValue{})
		if err != nil {
			t.Fatalf("%d, unexpected error", err)
		}

		var item map[string]interface{}
		err = dynamodbattribute.UnmarshalMap(resp.Attributes, &item)
		if err != nil {
			t.Fatalf("%d, unexpected error", err)
		}

		if item["foo"] != test.expected {
			t.Errorf("expected %+v, got %s", item["foo"], test.expected)
		}
	}
}

type mockedQuery struct {
	dynamodbiface.DynamoDBAPI
	Resp dynamodb.QueryOutput
}

func (m mockedQuery) QueryWithContext(ctx aws.Context, input *dynamodb.QueryInput, opts ...request.Option) (*dynamodb.QueryOutput, error) {
	return &m.Resp, nil
}

func TestQuery(t *testing.T) {
	tests := []struct {
		resp     dynamodb.QueryOutput
		expected string
	}{
		{
			resp: dynamodb.QueryOutput{
				Items: []map[string]*dynamodb.AttributeValue{
					{
						"foo": {
							S: aws.String("bar"),
						},
					},
				},
			},
			expected: "bar",
		},
	}

	ctx := context.TODO()

	for _, test := range tests {
		a := API{
			mockedQuery{Resp: test.resp},
		}

		resp, err := a.Query(ctx, "", "", "", "", 0, true)
		if err != nil {
			t.Fatalf("%d, unexpected error", err)
		}

		var items []map[string]interface{}
		for _, i := range resp.Items {
			var item map[string]interface{}
			err = dynamodbattribute.UnmarshalMap(i, &item)
			if err != nil {
				t.Fatalf("%d, unexpected error", err)
			}

			items = append(items, item)
		}

		if items[0]["foo"] != test.expected {
			t.Errorf("expected %+v, got %s", items[0]["foo"], test.expected)
		}
	}
}
