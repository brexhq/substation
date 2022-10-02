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

// errDynamoDBSinkJSON is returned when the DynamoDB sink receives non-JSON or invalid JSON data. If this error occurs, then parse the data into JSON or drop invalid JSON before it reaches the sink.
const errDynamoDBJSON = errors.Error("input must be JSON")

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
func (sink *DynamoDB) Send(ctx context.Context, ch *config.Channel) error {
	if !dynamodbAPI.IsEnabled() {
		dynamodbAPI.Setup()
	}

	var count int
	for cap := range ch.C {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if !json.Valid(cap.Data()) {
				return fmt.Errorf("sink dynamodb table %s: %v", sink.Table, errDynamoDBJSON)
			}

			items := cap.Get(sink.ItemsKey).Array()
			for _, item := range items {
				cache := make(map[string]interface{})
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
		"table", sink.Table,
	).WithField(
		"count", count,
	).Debug("put items into DynamoDB")

	return nil
}
