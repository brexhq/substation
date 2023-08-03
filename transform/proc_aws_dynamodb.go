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
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/json"
	mess "github.com/brexhq/substation/message"
)

// errProcAWSDynamodbInputNotAnObject is returned when the input is not an object.
// Refer to the dynamodb transform documentation for input requirements.
var errProcAWSDynamodbInputNotAnObject = fmt.Errorf("input is not an object")

// errProcAWSDynamodbInputMissingPK is returned when the JSON key "PK" is missing in
// the input. Refer to the dynamodb transform documentation for input requirements.
var errProcAWSDynamodbInputMissingPK = fmt.Errorf("input missing PK")

type procAWSDynamoDBConfig struct {
	Auth    config.ConfigAWSAuth `json:"auth"`
	Request config.ConfigRequest `json:"request"`
	// Key retrieves a value from an object for processing.
	//
	// This is optional for transforms that support processing non-object data.
	Key string `json:"key"`
	// SetKey inserts a processed value into an object.
	//
	// This is optional for transforms that support processing non-object data.
	SetKey string `json:"set_key"`
	// Table is the DynamoDB table that is queried.
	Table string `json:"table"`
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
	//
	// - true: traversal is performed in ascending order
	//
	// - false: traversal is performed in descending order
	//
	// This is optional and defaults to true.
	ScanIndexForward bool `json:"scan_index_forward"`
}

type procAWSDynamoDB struct {
	conf procAWSDynamoDBConfig

	// client is safe for concurrent access.
	client dynamodb.API
}

func newProcAWSDynamoDB(ctx context.Context, cfg config.Config) (*procAWSDynamoDB, error) {
	conf := procAWSDynamoDBConfig{}
	if err := config.Decode(cfg.Settings, &conf); err != nil {
		return nil, err
	}

	// Validate required options.
	if conf.Key == "" || conf.SetKey == "" {
		return nil, fmt.Errorf("transform: proc_aws_dynamodb: key %s set_key %s: %v", conf.Key, conf.SetKey, errInvalidDataPattern)
	}

	if conf.Table == "" {
		return nil, fmt.Errorf("transform: proc_aws_dynamodb: table: %v", errors.ErrMissingRequiredOption)
	}

	if conf.KeyConditionExpression == "" {
		return nil, fmt.Errorf("transform: proc_aws_dynamodb: key_condition_expression: %v", errors.ErrMissingRequiredOption)
	}

	proc := procAWSDynamoDB{
		conf: conf,
	}

	// Setup the AWS client.
	proc.client.Setup(aws.Config{
		Region:     conf.Auth.Region,
		AssumeRole: conf.Auth.AssumeRole,
		MaxRetries: conf.Request.MaxRetries,
	})

	return &proc, nil
}

func (t *procAWSDynamoDB) String() string {
	b, _ := gojson.Marshal(t.conf)
	return string(b)
}

func (*procAWSDynamoDB) Close(context.Context) error {
	return nil
}

func (t *procAWSDynamoDB) Transform(ctx context.Context, messages ...*mess.Message) ([]*mess.Message, error) {
	var output []*mess.Message

	for _, message := range messages {
		// Skip control messages.
		if message.IsControl() {
			output = append(output, message)
			continue
		}

		result := message.Get(t.conf.Key)
		if !result.IsObject() {
			return nil, fmt.Errorf("transform: proc_aws_dynamodb: key %s: %v", t.conf.Key, errProcAWSDynamodbInputNotAnObject)
		}

		// PK is a required field
		pk := json.Get([]byte(result.Raw), "PK").String()
		if pk == "" {
			return nil, fmt.Errorf("transform: proc_aws_dynamodb: key %s: %v", t.conf.Key, errProcAWSDynamodbInputMissingPK)
		}

		// SK is an optional field
		sk := json.Get([]byte(result.Raw), "SK").String()

		value, err := t.dynamodb(ctx, pk, sk)
		if err != nil {
			return nil, fmt.Errorf("transform: proc_aws_dynamodb: %v", err)
		}

		// no match
		if len(value) == 0 {
			output = append(output, message)
			continue
		}

		if err := message.Set(t.conf.SetKey, value); err != nil {
			return nil, fmt.Errorf("transform: proc_aws_dynamodb: %v", err)
		}

		output = append(output, message)
	}

	return output, nil
}

func (t *procAWSDynamoDB) dynamodb(ctx context.Context, pk, sk string) ([]map[string]interface{}, error) {
	resp, err := t.client.Query(
		ctx,
		t.conf.Table,
		pk, sk,
		t.conf.KeyConditionExpression,
		t.conf.Limit,
		t.conf.ScanIndexForward,
	)
	if err != nil {
		return nil, fmt.Errorf("transform: proc_aws_dynamodb: %v", err)
	}

	var items []map[string]interface{}
	for _, i := range resp.Items {
		var item map[string]interface{}
		err = dynamodbattribute.UnmarshalMap(i, &item)
		if err != nil {
			return nil, fmt.Errorf("transform: proc_aws_dynamodb: %v", err)
		}

		items = append(items, item)
	}
	return items, nil
}
