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

// awsDynamodb sinks JSON data to an AWS DynamoDB table. This sink supports writing multiple items from the same event to a table.
//
// This sink supports the object handling pattern.
type _awsDynamodb struct {
	// Table is the DynamoDB table that items are written to.
	Table string `json:"table"`
	// Key contains the DynamoDB items map that is written to the table.
	//
	// This supports one or more items by processing the key as an array.
	Key string `json:"key"`
}

// Send sinks a channel of encapsulated data with the sink.
func (sink *_awsDynamodb) Send(ctx context.Context, ch *config.Channel) error {
	if !dynamodbAPI.IsEnabled() {
		dynamodbAPI.Setup()
	}

	var count int
	for capsule := range ch.C {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if !json.Valid(capsule.Data()) {
				return fmt.Errorf("sink: aws_dynamodb: table %s: %v", sink.Table, errDynamoDBJSON)
			}

			items := capsule.Get(sink.Key).Array()
			for _, item := range items {
				cache := make(map[string]interface{})
				for k, v := range item.Map() {
					cache[k] = v.Value()
				}

				values, err := dynamodbattribute.MarshalMap(cache)
				if err != nil {
					return fmt.Errorf("sink: aws_dynamodb: table %s: %v", sink.Table, err)
				}

				_, err = dynamodbAPI.PutItem(ctx, sink.Table, values)
				if err != nil {
					// PutItem err returns metadata
					return fmt.Errorf("sink: aws_dynamodb: %v", err)
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
