package transform

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/aws"
	"github.com/brexhq/substation/internal/aws/dynamodb"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/message"
)

// errSendAWSDynamoDBNonObject is returned when non-object data is sent to the transform.
//
// If this error occurs, then parse the data into an object (or drop invalid objects)
// before attempting to send the data.
var errSendAWSDynamoDBNonObject = fmt.Errorf("input must be object")

type sendAWSDynamoDBConfig struct {
	Object iconfig.Object `json:"object"`
	AWS    iconfig.AWS    `json:"aws"`
	Retry  iconfig.Retry  `json:"retry"`

	// Table is the DynamoDB table that items are written to.
	Table string `json:"table"`
}

func (c *sendAWSDynamoDBConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *sendAWSDynamoDBConfig) Validate() error {
	if c.Object.Key == "" {
		return fmt.Errorf("object_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.Table == "" {
		return fmt.Errorf("table: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

type sendAWSDynamoDB struct {
	conf sendAWSDynamoDBConfig

	// client is safe for concurrent use.
	client dynamodb.API
}

func newSendAWSDynamoDB(_ context.Context, cfg config.Config) (*sendAWSDynamoDB, error) {
	conf := sendAWSDynamoDBConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: new_send_aws_dynamodb: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: new_send_aws_dynamodb: %v", err)
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

func (tf *sendAWSDynamoDB) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	if !json.Valid(msg.Data()) {
		return nil, fmt.Errorf("transform: send_aws_dynamodb: table %s: %v", tf.conf.Table, errSendAWSDynamoDBNonObject)
	}

	value := msg.GetValue(tf.conf.Object.Key)
	for _, item := range value.Array() {
		cache := make(map[string]interface{})
		for k, v := range item.Map() {
			cache[k] = v.Value()
		}

		attrVals, err := dynamodbattribute.MarshalMap(cache)
		if err != nil {
			return nil, fmt.Errorf("transform: send_aws_dynamodb: table %s: %v", tf.conf.Table, err)
		}

		if _, err = tf.client.PutItem(ctx, tf.conf.Table, attrVals); err != nil {
			// PutItem errors return metadata and don't require more information.
			return nil, fmt.Errorf("transform: send_aws_dynamodb: %v", err)
		}
	}

	return []*message.Message{msg}, nil
}

func (tf *sendAWSDynamoDB) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}

func (*sendAWSDynamoDB) Close(_ context.Context) error {
	return nil
}
