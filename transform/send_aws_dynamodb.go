package transform

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/aws"
	"github.com/brexhq/substation/internal/aws/dynamodb"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/json"
	mess "github.com/brexhq/substation/message"
)

// errSendDynamoDBNonObject is returned when non-object data is sent to the transform.
//
// If this error occurs, then parse the data into an object (or drop invalid objects)
// before attempting to send the data.
var errSendDynamoDBNonObject = fmt.Errorf("input must be object")

type sendAWSDynampDBConfig struct {
	Auth    config.ConfigAWSAuth `json:"auth"`
	Request config.ConfigRequest `json:"request"`
	// Table is the DynamoDB table that items are written to.
	Table string `json:"table"`
	// Key contains the DynamoDB items map that is written to the table.
	//
	// This supports one or more items by processing the key as an array.
	Key string `json:"key"`
}

type sendAWSDynamoDB struct {
	conf sendAWSDynampDBConfig

	// client is safe for concurrent use.
	client dynamodb.API
}

func newSendAWSDynamoDB(_ context.Context, cfg config.Config) (*sendAWSDynamoDB, error) {
	conf := sendAWSDynampDBConfig{}
	if err := config.Decode(cfg.Settings, &conf); err != nil {
		return nil, err
	}

	// Validate required options.
	if conf.Table == "" {
		return nil, fmt.Errorf("send: aws_kinesis: table: %v", errors.ErrMissingRequiredOption)
	}

	if conf.Key == "" {
		return nil, fmt.Errorf("send: aws_kinesis: key: %v", errors.ErrMissingRequiredOption)
	}

	send := sendAWSDynamoDB{
		conf: conf,
	}

	send.client.Setup(aws.Config{
		Region:     conf.Auth.Region,
		AssumeRole: conf.Auth.AssumeRole,
		MaxRetries: conf.Request.MaxRetries,
	})

	return &send, nil
}

func (*sendAWSDynamoDB) Close(_ context.Context) error {
	return nil
}

func (t *sendAWSDynamoDB) Transform(ctx context.Context, messages ...*mess.Message) ([]*mess.Message, error) {
	for _, message := range messages {
		if message.IsControl() {
			continue
		}

		if !json.Valid(message.Data()) {
			return nil, fmt.Errorf("send: aws_dynamodb: table %s: %v", t.conf.Table, errSendDynamoDBNonObject)
		}

		items := message.Get(t.conf.Key).Array()
		for _, item := range items {
			cache := make(map[string]interface{})
			for k, v := range item.Map() {
				cache[k] = v.Value()
			}

			values, err := dynamodbattribute.MarshalMap(cache)
			if err != nil {
				return nil, fmt.Errorf("send: aws_dynamodb: table %s: %v", t.conf.Table, err)
			}

			if _, err = t.client.PutItem(ctx, t.conf.Table, values); err != nil {
				// PutItem errors return metadata.
				return nil, fmt.Errorf("send: aws_dynamodb: %v", err)
			}
		}
	}

	return messages, nil
}
