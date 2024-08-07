package dynamodb

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/aws/aws-xray-sdk-go/xray"
	iaws "github.com/brexhq/substation/internal/aws"
)

// New returns a configured DynamoDB client.
func New(cfg iaws.Config) *dynamodb.DynamoDB {
	conf, sess := iaws.New(cfg)

	c := dynamodb.New(sess, conf)
	if _, ok := os.LookupEnv("AWS_XRAY_DAEMON_ADDRESS"); ok {
		xray.AWS(c.Client)
	}

	return c
}

// API wraps the DynamoDB API interface.
type API struct {
	Client dynamodbiface.DynamoDBAPI
}

// Setup creates a new DynamoDB client.
func (a *API) Setup(cfg iaws.Config) {
	a.Client = New(cfg)
}

// IsEnabled returns true if the client is enabled and ready for use.
func (a *API) IsEnabled() bool {
	return a.Client != nil
}

func (a *API) DeleteItem(ctx aws.Context, table string, key map[string]*dynamodb.AttributeValue) (resp *dynamodb.DeleteItemOutput, err error) {
	ctx = context.WithoutCancel(ctx)
	resp, err = a.Client.DeleteItemWithContext(
		ctx,
		&dynamodb.DeleteItemInput{
			TableName: aws.String(table),
			Key:       key,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("deleteitem table %s: %v", table, err)
	}

	return resp, nil
}

// BatchPutItem is a convenience wrapper for putting multiple items into a DynamoDB table.
func (a *API) BatchPutItem(ctx aws.Context, table string, items []map[string]*dynamodb.AttributeValue) (resp *dynamodb.BatchWriteItemOutput, err error) {
	var requests []*dynamodb.WriteRequest
	for _, item := range items {
		requests = append(requests, &dynamodb.WriteRequest{
			PutRequest: &dynamodb.PutRequest{
				Item: item,
			},
		})
	}

	ctx = context.WithoutCancel(ctx)
	resp, err = a.Client.BatchWriteItemWithContext(
		ctx,
		&dynamodb.BatchWriteItemInput{
			RequestItems: map[string][]*dynamodb.WriteRequest{
				table: requests,
			},
		},
	)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case dynamodb.ErrCodeProvisionedThroughputExceededException:
				var retry []map[string]*dynamodb.AttributeValue

				for _, item := range resp.UnprocessedItems[table] {
					retry = append(retry, item.PutRequest.Item)
				}

				if len(retry) > 0 {
					return a.BatchPutItem(ctx, table, retry)
				}

				fallthrough
			default:
				return nil, fmt.Errorf("batch_put_item: table %s: %v", table, err)
			}
		}
	}

	return resp, nil
}

// UpdateItem
func (a *API) UpdateItem(ctx aws.Context, input *dynamodb.UpdateItemInput) (resp *dynamodb.UpdateItemOutput, err error) {
	ctx = context.WithoutCancel(ctx)
	return a.Client.UpdateItemWithContext(ctx, input)
}

// PutItem is a convenience wrapper for putting items into a DynamoDB table.
func (a *API) PutItem(ctx aws.Context, table string, item map[string]*dynamodb.AttributeValue) (resp *dynamodb.PutItemOutput, err error) {
	ctx = context.WithoutCancel(ctx)
	resp, err = a.Client.PutItemWithContext(
		ctx,
		&dynamodb.PutItemInput{
			TableName: aws.String(table),
			Item:      item,
		})
	if err != nil {
		return nil, fmt.Errorf("putitem table %s: %v", table, err)
	}

	return resp, nil
}

func (a *API) PutItemWithCondition(ctx aws.Context, table string, item map[string]*dynamodb.AttributeValue, conditionExpression string, expressionAttributeNames map[string]*string, expressionAttributeValues map[string]*dynamodb.AttributeValue) (resp *dynamodb.PutItemOutput, err error) {
	input := &dynamodb.PutItemInput{
		TableName:                 aws.String(table),
		ConditionExpression:       aws.String(conditionExpression),
		ExpressionAttributeNames:  expressionAttributeNames,
		Item:                      item,
		ExpressionAttributeValues: expressionAttributeValues,
		ReturnValues:              aws.String("ALL_OLD"),
	}

	resp, err = a.Client.PutItemWithContext(ctx, input)
	if err != nil {
		return resp, err
	}

	return resp, nil
}

/*
Query is a convenience wrapper for querying a DynamoDB table. The paritition and sort keys are always referenced in the key condition expression as ":PK" and ":SK". Refer to the DynamoDB documentation for the Query operation's request syntax and key condition expression patterns:

- https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_Query.html#API_Query_RequestSyntax

- https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Query.html#Query.KeyConditionExpressions
*/
func (a *API) Query(ctx aws.Context, table, partitionKey, sortKey, keyConditionExpression string, limit int64, scanIndexForward bool) (resp *dynamodb.QueryOutput, err error) {
	expression := make(map[string]*dynamodb.AttributeValue)
	expression[":PK"] = &dynamodb.AttributeValue{
		S: aws.String(partitionKey),
	}

	if sortKey != "" {
		expression[":SK"] = &dynamodb.AttributeValue{
			S: aws.String(sortKey),
		}
	}

	ctx = context.WithoutCancel(ctx)
	resp, err = a.Client.QueryWithContext(
		ctx,
		&dynamodb.QueryInput{
			TableName:                 aws.String(table),
			KeyConditionExpression:    aws.String(keyConditionExpression),
			ExpressionAttributeValues: expression,
			Limit:                     aws.Int64(limit),
			ScanIndexForward:          aws.Bool(scanIndexForward),
		})
	if err != nil {
		return nil, fmt.Errorf("query: table %s key_condition_expression %s: %v", table, keyConditionExpression, err)
	}

	return resp, nil
}

// GetItem is a convenience wrapper for getting items into a DynamoDB table.
func (a *API) GetItem(ctx aws.Context, table string, attributes map[string]interface{}, consistentRead bool) (resp *dynamodb.GetItemOutput, err error) {
	attr, err := dynamodbattribute.MarshalMap(attributes)
	if err != nil {
		return nil, fmt.Errorf("get_item: table %s: %v", table, err)
	}

	ctx = context.WithoutCancel(ctx)
	resp, err = a.Client.GetItemWithContext(
		ctx,
		&dynamodb.GetItemInput{
			TableName:      aws.String(table),
			Key:            attr,
			ConsistentRead: aws.Bool(consistentRead),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("get_item: table %s: %v", table, err)
	}

	return resp, nil
}

// ConvertEventsAttributeValue converts events.DynamoDBAttributeValue to dynamodb.AttributeValue.
func ConvertEventsAttributeValue(v events.DynamoDBAttributeValue) *dynamodb.AttributeValue {
	switch v.DataType() {
	case events.DataTypeBinary:
		return &dynamodb.AttributeValue{
			B: v.Binary(),
		}
	case events.DataTypeBinarySet:
		return &dynamodb.AttributeValue{
			BS: v.BinarySet(),
		}
	case events.DataTypeNumber:
		return &dynamodb.AttributeValue{
			N: aws.String(v.Number()),
		}
	case events.DataTypeNumberSet:
		av := &dynamodb.AttributeValue{}

		for _, n := range v.NumberSet() {
			av.NS = append(av.NS, aws.String(n))
		}

		return av
	case events.DataTypeString:
		return &dynamodb.AttributeValue{
			S: aws.String(v.String()),
		}
	case events.DataTypeStringSet:
		av := &dynamodb.AttributeValue{}

		for _, s := range v.StringSet() {
			av.SS = append(av.SS, aws.String(s))
		}

		return av
	case events.DataTypeList:
		av := &dynamodb.AttributeValue{}

		for _, v := range v.List() {
			av.L = append(av.L, ConvertEventsAttributeValue(v))
		}

		return av
	case events.DataTypeMap:
		av := &dynamodb.AttributeValue{}
		av.M = make(map[string]*dynamodb.AttributeValue)

		for k, v := range v.Map() {
			av.M[k] = ConvertEventsAttributeValue(v)
		}

		return av
	case events.DataTypeNull:
		return &dynamodb.AttributeValue{
			NULL: aws.Bool(true),
		}
	case events.DataTypeBoolean:
		return &dynamodb.AttributeValue{
			BOOL: aws.Bool(v.Boolean()),
		}
	default:
		return nil
	}
}

// ConvertEventsAttributeValueMap converts a map of events.DynamoDBAttributeValue to a map of dynamodb.AttributeValue.
func ConvertEventsAttributeValueMap(m map[string]events.DynamoDBAttributeValue) map[string]*dynamodb.AttributeValue {
	av := make(map[string]*dynamodb.AttributeValue)

	for k, v := range m {
		av[k] = ConvertEventsAttributeValue(v)
	}

	return av
}
