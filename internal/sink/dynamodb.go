package sink

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"

	"github.com/brexhq/substation/internal/aws/dynamodb"
	"github.com/brexhq/substation/internal/json"
	"github.com/brexhq/substation/internal/log"
)

/*
DynamoDB sinks JSON data to AWS DynamoDB tables. This sink supports sinking multiple rows from the same event to a DynamoDB table.

The sink has these settings:
	Table:
		DynamoDB table that data is written to
	Attributes:
		maps values from the JSON object (Key) to attributes in the DynamoDB table (Attribute)
	ErrorOnFailure (optional):
		if set to true, then receiving non-JSON data will cause the sink to fail
		defaults to false

The sink uses this Jsonnet configuration:
	{
		type: 'dynamodb',
		settings: {
			table: 'foo-table',
			attributes: [
				[
					{
						key: 'foo',
						attribute: 'bar',
					},
				],
			],
		},
	}
*/
type DynamoDB struct {
	Table      string `json:"table"`
	Attributes [][]struct {
		Key       string `json:"key"`
		Attribute string `json:"attribute"`
	} `json:"attributes"`
	ErrorOnFailure bool `json:"error_on_failure"`
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
			if !json.Valid(data) && sink.ErrorOnFailure {
				return fmt.Errorf("err DynamoDB sink received invalid JSON data: %v", json.JSONInvalidData)
			} else if !json.Valid(data) {
				log.Info("DynamoDB sink received invalid JSON data")
				continue
			}

			for _, attr := range sink.Attributes {
				var cache map[string]interface{}
				cache = make(map[string]interface{})
				for _, field := range attr {
					cache[field.Attribute] = json.Get(data, field.Key).Value()
				}

				// if cache is empty, then all match condition failed
				if len(cache) == 0 {
					continue
				}

				values, err := dynamodbattribute.MarshalMap(cache)
				if err != nil {
					return fmt.Errorf("err marshalling DynamoDB results: %v", err)
				}

				_, err = dynamodbAPI.PutItem(ctx, sink.Table, values)
				if err != nil {
					return fmt.Errorf("err putting values into DynamoDB table %s: %v", sink.Table, err)
				}
			}

			count++
		}
	}

	log.WithField(
		"count", count,
	).Debug("put items into DynamoDB")

	return nil
}
