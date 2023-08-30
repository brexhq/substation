package transform

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/aws"
	"github.com/brexhq/substation/internal/aws/dynamodb"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/json"
	"github.com/brexhq/substation/message"
)

// errSendDynamoDBNonObject is returned when non-object data is sent to the transform.
//
// If this error occurs, then parse the data into an object (or drop invalid objects)
// before attempting to send the data.
var errSendDynamoDBNonObject = fmt.Errorf("input must be object")

type sendAWSDynampDBConfig struct {
	AWS   configAWS   `json:"aws"`
	Retry configRetry `json:"retry"`

	// Table is the DynamoDB table that items are written to.
	Table string `json:"table"`
	// Key contains the DynamoDB items that are written to the table.
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
	if err := iconfig.Decode(cfg.Settings, &conf); err != nil {
		return nil, fmt.Errorf("transform: new_send_aws_dynamodb: %v", err)
	}

	// Validate required options.
	if conf.Table == "" {
		return nil, fmt.Errorf("transform: new_send_aws_dynamodb: table: %v", errors.ErrMissingRequiredOption)
	}

	if conf.Key == "" {
		return nil, fmt.Errorf("transform: new_send_aws_dynamodb: obj_get_key: %v", errors.ErrMissingRequiredOption)
	}

	tf := sendAWSDynamoDB{
		conf: conf,
	}

	tf.client.Setup(aws.Config{
		Region:     conf.AWS.Region,
		AssumeRole: conf.AWS.AssumeRole,
		MaxRetries: conf.Retry.Attempts,
	})

	return &tf, nil
}

func (*sendAWSDynamoDB) Close(_ context.Context) error {
	return nil
}

func (tf *sendAWSDynamoDB) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	if !json.Valid(msg.Data()) {
		return nil, fmt.Errorf("send: aws_dynamodb: table %s: %v", tf.conf.Table, errSendDynamoDBNonObject)
	}

	items := msg.GetObject(tf.conf.Key).Array()
	for _, item := range items {
		cache := make(map[string]interface{})
		for k, v := range item.Map() {
			cache[k] = v.Value()
		}

		values, err := dynamodbattribute.MarshalMap(cache)
		if err != nil {
			return nil, fmt.Errorf("send: aws_dynamodb: table %s: %v", tf.conf.Table, err)
		}

		if _, err = tf.client.PutItem(ctx, tf.conf.Table, values); err != nil {
			// PutItem errors return metadata.
			return nil, fmt.Errorf("send: aws_dynamodb: %v", err)
		}
	}

	return []*message.Message{msg}, nil
}
