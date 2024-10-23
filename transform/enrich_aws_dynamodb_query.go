package transform

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"

	iconfig "github.com/brexhq/substation/v2/internal/config"
)

type enrichAWSDynamoDBQueryQueryConfig struct {
	Attributes struct {
		// PartitionKey is the table's parition key attribute.
		//
		// This is required for all tables.
		PartitionKey string `json:"partition_key"`
		// SortKey is the table's sort (range) key attribute.
		//
		// This must be used if the table uses a composite primary key schema
		// (partition key and sort key). Only string types are supported.
		SortKey string `json:"sort_key"`
	} `json:"attributes"`
	// Limits determines the maximum number of items to evalute.
	//
	// This is optional and defaults to evaluating all items.
	Limit int32 `json:"limit"`
	// ScanIndexForward specifies the order of index traversal.
	//
	// Must be one of:
	//	- true (traversal is performed in ascending order)
	//	- false (traversal is performed in descending order)
	//
	// This is optional and defaults to true.
	ScanIndexForward bool `json:"scan_index_forward"`

	ID     string         `json:"id"`
	Object iconfig.Object `json:"object"`
	AWS    iconfig.AWS    `json:"aws"`
}

func (c *enrichAWSDynamoDBQueryQueryConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *enrichAWSDynamoDBQueryQueryConfig) Validate() error {
	if c.Object.TargetKey == "" {
		return fmt.Errorf("object.target_key: %v", iconfig.ErrMissingRequiredOption)
	}

	if c.Attributes.PartitionKey == "" {
		return fmt.Errorf("attributes.partition_key: %v", iconfig.ErrMissingRequiredOption)
	}

	if c.AWS.ARN == "" {
		return fmt.Errorf("aws.arn: %v", iconfig.ErrMissingRequiredOption)
	}

	return nil
}

func newEnrichAWSDynamoDBQuery(ctx context.Context, cfg config.Config) (*enrichAWSDynamoDBQuery, error) {
	conf := enrichAWSDynamoDBQueryQueryConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform enrich_aws_dynamodb_query: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "enrich_aws_dynamodb_query"
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	tf := enrichAWSDynamoDBQuery{
		conf: conf,
	}

	awsCfg, err := iconfig.NewAWS(ctx, conf.AWS)
	if err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	tf.client = dynamodb.NewFromConfig(awsCfg)

	return &tf, nil
}

type enrichAWSDynamoDBQuery struct {
	conf   enrichAWSDynamoDBQueryQueryConfig
	client *dynamodb.Client
}

func (tf *enrichAWSDynamoDBQuery) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.HasFlag(message.IsControl) {
		return []*message.Message{msg}, nil
	}

	value := msg.GetValue(tf.conf.Object.SourceKey)
	if skipMessage(msg, value) {
		return []*message.Message{msg}, nil
	}

	// This supports one of two states:
	//	- A single partition key, captured as a string.
	//	- A composite key (partition key and sort key), captured as an array of two strings.
	//
	// If the value is an array, we assume it is a composite key.
	var keyEx expression.KeyConditionBuilder
	if value.IsArray() && len(value.Array()) == 2 && tf.conf.Attributes.SortKey != "" {
		keyEx = expression.Key(tf.conf.Attributes.PartitionKey).Equal(expression.Value(value.Array()[0].String())).
			And(expression.Key(tf.conf.Attributes.SortKey).Equal(expression.Value(value.Array()[1].String())))
	} else if !value.IsArray() {
		keyEx = expression.Key(tf.conf.Attributes.PartitionKey).Equal(expression.Value(value.String()))
	} else { // This is invalid, so we return the original message.
		return []*message.Message{msg}, nil
	}

	expr, err := expression.NewBuilder().WithKeyCondition(keyEx).Build()
	if err != nil {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
	}

	ctx = context.WithoutCancel(ctx)
	resp, err := tf.client.Query(ctx, &dynamodb.QueryInput{
		TableName:                 &tf.conf.AWS.ARN,
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		Limit:                     aws.Int32(tf.conf.Limit),
		ScanIndexForward:          aws.Bool(tf.conf.ScanIndexForward),
	})
	if err != nil {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
	}

	var items []map[string]interface{}
	for _, i := range resp.Items {
		var item map[string]interface{}
		if err := attributevalue.UnmarshalMap(i, &item); err != nil {
			return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
		}

		items = append(items, item)
	}

	if len(items) == 0 {
		return []*message.Message{msg}, nil
	}

	if err := msg.SetValue(tf.conf.Object.TargetKey, items); err != nil {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
	}

	return []*message.Message{msg}, nil
}

func (tf *enrichAWSDynamoDBQuery) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
