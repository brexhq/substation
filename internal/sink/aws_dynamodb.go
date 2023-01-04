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

// errDynamoDBNonObject is returned when the DynamoDB sink receives non-object data.
//
// If this error occurs, then parse the data into an object (or drop invalid objects)
// before it reaches the sink.
const errDynamoDBNonObject = errors.Error("input must be object")

// awsDynamodb sinks data to an AWS DynamoDB table.
//
// Writing multiple items from the same object to a table is possible when the
// input is an array of item payloads.
type sinkAWSDynamoDB struct {
	// Table is the DynamoDB table that items are written to.
	Table string `json:"table"`
	// Key contains the DynamoDB items map that is written to the table.
	//
	// This supports one or more items by processing the key as an array.
	Key string `json:"key"`
}

// Send sinks a channel of encapsulated data with the sink.
func (s *sinkAWSDynamoDB) Send(ctx context.Context, ch *config.Channel) error {
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
				return fmt.Errorf("sink: aws_dynamodb: table %s: %v", s.Table, errDynamoDBNonObject)
			}

			items := capsule.Get(s.Key).Array()
			for _, item := range items {
				cache := make(map[string]interface{})
				for k, v := range item.Map() {
					cache[k] = v.Value()
				}

				values, err := dynamodbattribute.MarshalMap(cache)
				if err != nil {
					return fmt.Errorf("sink: aws_dynamodb: table %s: %v", s.Table, err)
				}

				_, err = dynamodbAPI.PutItem(ctx, s.Table, values)
				if err != nil {
					// PutItem err returns metadata
					return fmt.Errorf("sink: aws_dynamodb: %v", err)
				}

				count++
			}
		}
	}

	log.WithField(
		"table", s.Table,
	).WithField(
		"count", count,
	).Debug("put items into DynamoDB")

	return nil
}
