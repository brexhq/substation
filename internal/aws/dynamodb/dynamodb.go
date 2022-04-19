package dynamodb

import (
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/aws/aws-xray-sdk-go/xray"
)

//New creates and returns a new session connection to DynamoDB
func New() *dynamodb.DynamoDB {
	conf := aws.NewConfig()

	// provides forward compatibility for the Go SDK to support env var configuration settings
	// https://github.com/aws/aws-sdk-go/issues/4207
	max, found := os.LookupEnv("AWS_MAX_ATTEMPTS")
	if found {
		m, err := strconv.Atoi(max)
		if err != nil {
			panic(err)
		}

		conf = conf.WithMaxRetries(m)
	}

	c := dynamodb.New(
		session.Must(session.NewSession()),
		conf,
	)
	xray.AWS(c.Client)
	return c
}

// API wraps a DynamoDB client interface
type API struct {
	Client dynamodbiface.DynamoDBAPI
}

// Setup creates a DynamoDB client and sets the DynamoDB.stream
func (a *API) Setup() {
	a.Client = New()
}

// IsEnabled checks whether a new client has been set
func (a *API) IsEnabled() bool {
	return a.Client != nil
}

// PutItem is a convenience wrapper for executing the PutItem API on DynamoDB.table
func (a *API) PutItem(ctx aws.Context, table string, item map[string]*dynamodb.AttributeValue) (resp *dynamodb.PutItemOutput, err error) {
	resp, err = a.Client.PutItemWithContext(
		ctx,
		&dynamodb.PutItemInput{
			TableName: aws.String(table),
			Item:      item,
		})

	if err != nil {
		return resp, err
	}

	return resp, nil
}

// Query is a convenience wrapper for querying a DynamoDB table
func (a *API) Query(ctx aws.Context, table, partitionKey, sortKey, keyConditionExpression string, limit int64, scanIndexForward bool) (resp *dynamodb.QueryOutput, err error) {
	var expression map[string]*dynamodb.AttributeValue

	if len(sortKey) != 0 {
		expression = map[string]*dynamodb.AttributeValue{
			":partitionkeyval": {
				S: aws.String(partitionKey),
			},
			":sortkeyval": {
				S: aws.String(sortKey),
			},
		}
	} else {
		expression = map[string]*dynamodb.AttributeValue{
			":partitionkeyval": {
				S: aws.String(partitionKey),
			},
		}
	}

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
		return resp, err
	}

	return resp, nil
}
