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

type enrichAWSDynamoDBConfig struct {
	Object iconfig.Object `json:"object"`
	AWS    iconfig.AWS    `json:"aws"`
	Retry  iconfig.Retry  `json:"retry"`

	// TableName is the DynamoDB table that is queried.
	TableName string `json:"table_name"`
	// PartitionKey is the DynamoDB partition key.
	PartitionKey string `json:"partition_key"`
	// SortKey is the DynamoDB sort key.
	SortKey string `json:"sort_key"`
	// KeyConditionExpression is the DynamoDB key condition
	// expression string (see documentation).
	KeyConditionExpression string `json:"key_condition_expression"`
	// Limits determines the maximum number of items to evalute.
	//
	// This is optional and defaults to evaluating all items.
	Limit int64 `json:"limit"`
	// ScanIndexForward specifies the order of index traversal.
	//
	// Must be one of:
	//	- true (traversal is performed in ascending order)
	//	- false (traversal is performed in descending order)
	//
	// This is optional and defaults to true.
	ScanIndexForward bool `json:"scan_index_forward"`
}

func (c *enrichAWSDynamoDBConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *enrichAWSDynamoDBConfig) Validate() error {
	if c.Object.SetKey == "" {
		return fmt.Errorf("object_set_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.PartitionKey == "" {
		return fmt.Errorf("partition_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.TableName == "" {
		return fmt.Errorf("table_name: %v", errors.ErrMissingRequiredOption)
	}

	if c.KeyConditionExpression == "" {
		return fmt.Errorf("key_condition_expression: %v", errors.ErrMissingRequiredOption)
	}
	return nil
}

func newEnrichAWSDynamoDB(_ context.Context, cfg config.Config) (*enrichAWSDynamoDB, error) {
	conf := enrichAWSDynamoDBConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: new_enrich_aws_dynamodb: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: new_enrich_aws_dynamodb: %v", err)
	}

	tf := enrichAWSDynamoDB{
		conf: conf,
	}

	// Setup the AWS client.
	tf.client.Setup(aws.Config{
		Region:        conf.AWS.Region,
		AssumeRoleARN: conf.AWS.AssumeRoleARN,
		MaxRetries:    conf.Retry.Count,
	})

	return &tf, nil
}

type enrichAWSDynamoDB struct {
	conf enrichAWSDynamoDBConfig

	// client is safe for concurrent access.
	client dynamodb.API
}

func (tf *enrichAWSDynamoDB) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	pk := msg.GetValue(tf.conf.PartitionKey).String()
	sk := msg.GetValue(tf.conf.SortKey).String()

	value, err := tf.dynamodb(ctx, pk, sk)
	if err != nil {
		return nil, fmt.Errorf("transform: enrich_aws_dynamodb: %v", err)
	}

	// No match.
	if len(value) == 0 {
		return []*message.Message{msg}, nil
	}

	if err := msg.SetValue(tf.conf.Object.SetKey, value); err != nil {
		return nil, fmt.Errorf("transform: enrich_aws_dynamodb: %v", err)
	}

	return []*message.Message{msg}, nil
}

func (tf *enrichAWSDynamoDB) dynamodb(ctx context.Context, pk, sk string) ([]map[string]interface{}, error) {
	resp, err := tf.client.Query(
		ctx,
		tf.conf.TableName,
		pk, sk,
		tf.conf.KeyConditionExpression,
		tf.conf.Limit,
		tf.conf.ScanIndexForward,
	)
	if err != nil {
		return nil, err
	}

	var items []map[string]interface{}
	for _, i := range resp.Items {
		var item map[string]interface{}
		err = dynamodbattribute.UnmarshalMap(i, &item)
		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}
	return items, nil
}

func (tf *enrichAWSDynamoDB) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
