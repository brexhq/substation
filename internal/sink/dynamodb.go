package sink

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/internal/aws/dynamodb"
	"github.com/brexhq/substation/internal/json"
	"github.com/brexhq/substation/internal/log"
)

/*
DynamoDB implements the Sink interface and writes data to DynamoDB tables. More information is available in the README.

Items: array of options for inspecting input data and putting an item into DynamoDB
Items.Condition: conditions that must pass to put data into DynamoDB
Items.Table: the DynamoDB table that data is written to
Items.Fields: maps keys from JSON data to a DynamoDB attribute / column
ErrorOnFailure (optional): determines if invalid input data causes the sink to error; defaults to false
*/
type DynamoDB struct {
	api   dynamodb.API
	Items []struct {
		Condition condition.OperatorConfig `mapstructure:"condition"`
		Table     string                   `mapstructure:"table"`
		Fields    []struct {
			Key       string `mapstructure:"key"`
			Attribute string `mapstructure:"attribute"`
		} `mapstructure:"fields"`
	} `mapstructure:"items"`
	ErrorOnFailure bool `mapstructure:"error_on_failure"`
}

// Send sends a channel of bytes to the DynamoDB tables defined by this sink.
func (sink *DynamoDB) Send(ctx context.Context, ch chan []byte, kill chan struct{}) error {
	if !sink.api.IsEnabled() {
		sink.api.Setup()
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

			var table string
			var cache map[string]interface{}

			for _, attributes := range sink.Items {
				op, err := condition.OperatorFactory(attributes.Condition)
				if err != nil {
					return err
				}
				ok, err := op.Operate(data)
				if err != nil {
					return err
				}

				if !ok {
					continue
				}

				table = attributes.Table
				cache = make(map[string]interface{})
				for _, field := range attributes.Fields {
					cache[field.Attribute] = json.Get(data, field.Key).Value()
				}
			}

			// if cache is empty, then all match condition failed
			if len(cache) == 0 {
				continue
			}

			values, err := dynamodbattribute.MarshalMap(cache)
			if err != nil {
				return fmt.Errorf("err marshalling DynamoDB results: %v", err)
			}

			_, err = sink.api.PutItem(ctx, table, values)
			if err != nil {
				return fmt.Errorf("err putting values into DyanmoDB table %s: %v", table, err)
			}

			count++
		}
	}

	log.WithField(
		"count", count,
	).Debug("put items into DynamoDB")

	return nil
}
