package sink

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/aws/dynamodb"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/json"
	"github.com/brexhq/substation/internal/log"
)

var dynamodbAPI dynamodb.API

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

When loaded with a factory, the sink uses this JSON configuration:
	{
		"type": "dynamodb",
		"settings": {
			"table": "foo-table",
			"items_key": "foo"
		}
	}
*/
type DynamoDB struct {
	Table    string `json:"table"`
	ItemsKey string `json:"items_key"`
}

// Send sinks a channel of encapsulated data with the DynamoDB sink.
func (sink *DynamoDB) Send(ctx context.Context, ch chan config.Capsule, kill chan struct{}) error {
	if !dynamodbAPI.IsEnabled() {
		dynamodbAPI.Setup()
	}

	var count int
	for cap := range ch {
		select {
		case <-kill:
			return nil
		default:
			if !json.Valid(cap.GetData()) {
				return fmt.Errorf("sink dynamodb table %s: %v", sink.Table, DynamoDBSinkInvalidJSON)
			}

			items := cap.Get(sink.ItemsKey).Array()
			for _, item := range items {
				var cache map[string]interface{}
				cache = make(map[string]interface{})
				for k, v := range item.Map() {
					cache[k] = v.Value()
				}

				values, err := dynamodbattribute.MarshalMap(cache)
				if err != nil {
					return fmt.Errorf("sink dynamodb table %s: %v", sink.Table, err)
				}

				_, err = dynamodbAPI.PutItem(ctx, sink.Table, values)
				if err != nil {
					// PutItem err returns metadata
					return fmt.Errorf("sink dynamodb: %v", err)
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
