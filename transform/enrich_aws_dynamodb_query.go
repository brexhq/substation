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

	iaws "github.com/brexhq/substation/v2/internal/aws"
	iconfig "github.com/brexhq/substation/v2/internal/config"
	ierrors "github.com/brexhq/substation/v2/internal/errors"
)

type enrichAWSDynamoDBQueryQueryConfig struct {
	// PartitionKey is the DynamoDB partition key.
	PartitionKey string `json:"partition_key"`
	// SortKey is the DynamoDB sort key.
	//
	// This is optional and has no default.
	SortKey string `json:"sort_key"`
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
		return fmt.Errorf("object_target_key: %v", ierrors.ErrMissingRequiredOption)
	}

	if c.PartitionKey == "" {
		return fmt.Errorf("partition_key: %v", ierrors.ErrMissingRequiredOption)
	}

	if c.AWS.ARN == "" {
		return fmt.Errorf("aws.arn: %v", ierrors.ErrMissingRequiredOption)
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

	awsCfg, err := iaws.New(ctx, iaws.Config{
		Region:  iaws.ParseRegion(conf.AWS.ARN),
		RoleARN: conf.AWS.AssumeRoleARN,
	})
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
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	value := msg.GetValue(tf.conf.Object.SourceKey)
	if !value.Exists() {
		return []*message.Message{msg}, nil
	}

	var PK, SK string
	if value.IsArray() {
		if len(value.Array()) != 2 {
			return nil, fmt.Errorf("transform %s: expected array of 2 elements, got %d", tf.conf.ID, len(value.Array()))
		}

		PK = value.Array()[0].String()
		SK = value.Array()[1].String()
	} else {
		PK = value.String()
	}

	keyEx := expression.Key(tf.conf.PartitionKey).Equal(expression.Value(PK))
	if tf.conf.SortKey != "" {
		keyEx = keyEx.And(expression.Key(tf.conf.SortKey).Equal(expression.Value(SK)))
	}

	expr, err := expression.NewBuilder().WithKeyCondition(keyEx).Build()
	if err != nil {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
	}
	input := &dynamodb.QueryInput{
		TableName:                 &tf.conf.AWS.ARN,
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		Limit:                     aws.Int32(tf.conf.Limit),
		ScanIndexForward:          aws.Bool(tf.conf.ScanIndexForward),
	}

	resp, err := tf.client.Query(ctx, input)
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
