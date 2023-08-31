//go:build !wasm

package transform

import (
	"context"
	gojson "encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/aws"
	"github.com/brexhq/substation/internal/aws/dynamodb"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/message"
)

type modAWSDynamoDBConfig struct {
	Object configObject `json:"object"`
	AWS    configAWS    `json:"aws"`
	Retry  configRetry  `json:"retry"`

	// Table is the DynamoDB table that is queried.
	Table string `json:"table"`
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

type modAWSDynamoDB struct {
	conf modAWSDynamoDBConfig

	// client is safe for concurrent access.
	client dynamodb.API
}

func newModAWSDynamoDB(_ context.Context, cfg config.Config) (*modAWSDynamoDB, error) {
	conf := modAWSDynamoDBConfig{}
	if err := iconfig.Decode(cfg.Settings, &conf); err != nil {
		return nil, fmt.Errorf("transform: new_mod_aws_dynamodb: %v", err)
	}

	// Validate required options.
	if conf.Object.SetKey == "" {
		return nil, fmt.Errorf("transform: new_mod_aws_dynamodb: object_set_key: %v", errors.ErrMissingRequiredOption)
	}

	if conf.PartitionKey == "" {
		return nil, fmt.Errorf("transform: new_mod_aws_dynamodb: partition_key: %v", errors.ErrMissingRequiredOption)
	}

	if conf.Table == "" {
		return nil, fmt.Errorf("transform: new_mod_aws_dynamodb: table: %v", errors.ErrMissingRequiredOption)
	}

	if conf.KeyConditionExpression == "" {
		return nil, fmt.Errorf("transform: new_mod_aws_dynamodb: key_condition_expression: %v", errors.ErrMissingRequiredOption)
	}

	tf := modAWSDynamoDB{
		conf: conf,
	}

	// Setup the AWS client.
	tf.client.Setup(aws.Config{
		Region:     conf.AWS.Region,
		AssumeRole: conf.AWS.AssumeRole,
		MaxRetries: conf.Retry.Attempts,
	})

	return &tf, nil
}

func (tf *modAWSDynamoDB) String() string {
	b, _ := gojson.Marshal(tf.conf)
	return string(b)
}

func (*modAWSDynamoDB) Close(context.Context) error {
	return nil
}

func (tf *modAWSDynamoDB) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	// Skip interrupt messages.
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	pk := msg.GetObject(tf.conf.PartitionKey).String()
	sk := msg.GetObject(tf.conf.SortKey).String()

	value, err := tf.dynamodb(ctx, pk, sk)
	if err != nil {
		return nil, fmt.Errorf("transform: mod_aws_dynamodb: %v", err)
	}

	// No match.
	if len(value) == 0 {
		return []*message.Message{msg}, nil
	}

	if err := msg.SetObject(tf.conf.Object.SetKey, value); err != nil {
		return nil, fmt.Errorf("transform: mod_aws_dynamodb: %v", err)
	}

	return []*message.Message{msg}, nil
}

func (tf *modAWSDynamoDB) dynamodb(ctx context.Context, pk, sk string) ([]map[string]interface{}, error) {
	resp, err := tf.client.Query(
		ctx,
		tf.conf.Table,
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
