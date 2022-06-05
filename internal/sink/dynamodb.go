package sink

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/brexhq/substation/internal/aws/dynamodb"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/json"
	"github.com/brexhq/substation/internal/log"
)

// DynamoDBSinkInvalidJSON is returned when the DynamoDB sink receives invalid JSON. If this error occurs, then parse the data into valid JSON or drop invalid JSON before it reaches the sink.
const DynamoDBSinkInvalidJSON = errors.Error("DynamoDBSinkInvalidJSON")

/*
DynamoDB sinks JSON data to an AWS DynamoDB table. This sink supports sinking multiple rows from the same event to a table.

The sink has these settings:
	Table:
		DynamoDB table that data is written to
	ItemsKey:
		JSON key-value that contains maps that represent items to be stored in the DynamoDB table
		This key can be a single map or an array of maps:
			[
				{
					"PK": "foo",
					"SK": "bar",
				},
				{
					"PK": "baz",
					"SK": "qux",
				}
			]

The sink uses this Jsonnet configuration:
	{
		type: 'dynamodb',
		settings: {
			table: 'foo-table',
			items_key: 'foo',
		},
	}
*/
type DynamoDB struct {
	Table    string `json:"table"`
	ItemsKey string `json:"items_key"`
}

var dynamodbAPI dynamodb.API

// Send sinks a channel of bytes with the DynamoDB sink.
func (sink *DynamoDB) Send(ctx context.Context, ch chan []byte, kill chan struct{}) error {
	if !dynamodbAPI.IsEnabled() {
		dynamodbAPI.Setup()
	}

	var count int
	for data := range ch {
		select {
		case <-kill:
			return nil
		default:
			// can only parse valid JSON into DynamoDB attributes
			// if this error occurs, then parse the data into JSON
			if !json.Valid(data) {
				log.Info("DynamoDB sink received invalid JSON data")
				return DynamoDBSinkInvalidJSON
			}

			items := json.Get(data, sink.ItemsKey).Array()
			for _, item := range items {
				var cache map[string]interface{}
				cache = make(map[string]interface{})
				for k, v := range item.Map() {
					cache[k] = v.Value()
				}

				values, err := dynamodbattribute.MarshalMap(cache)
				if err != nil {
					return fmt.Errorf("err marshalling DynamoDB results: %v", err)
				}

				_, err = dynamodbAPI.PutItem(ctx, sink.Table, values)
				if err != nil {
					return fmt.Errorf("err putting values into DynamoDB table %s: %v", sink.Table, err)
				}
				count++
			}
		}
	}

	log.WithField(
		"count", count,
	).WithField(
		"table", sink.Table,
	).Debug("put items into DynamoDB")

	return nil
}
