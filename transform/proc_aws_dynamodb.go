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
	_config "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/json"
	mess "github.com/brexhq/substation/message"
)

// errprocAWSDynamoDBInputNotAnObject is returned when the input is not an object.
// Refer to the dynamodb transform documentation for input requirements.
var errprocAWSDynamoDBInputNotAnObject = fmt.Errorf("input is not an object")

// errprocAWSDynamoDBInputMissingPK is returned when the JSON key "PK" is missing in
// the input. Refer to the dynamodb transform documentation for input requirements.
var errprocAWSDynamoDBInputMissingPK = fmt.Errorf("input missing PK")

type procAWSDynamoDBConfig struct {
	Auth    _config.ConfigAWSAuth `json:"auth"`
	Request _config.ConfigRequest `json:"request"`
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
	if err := _config.Decode(cfg.Settings, &conf); err != nil {
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

	mod := procAWSDynamoDB{
		conf: conf,
	}

	// Setup the AWS client.
	mod.client.Setup(aws.Config{
		Region:     conf.Auth.Region,
		AssumeRole: conf.Auth.AssumeRole,
		MaxRetries: conf.Request.MaxRetries,
	})

	return &mod, nil
}

func (mod *procAWSDynamoDB) String() string {
	b, _ := gojson.Marshal(mod.conf)
	return string(b)
}

func (*procAWSDynamoDB) Close(context.Context) error {
	return nil
}

func (mod *procAWSDynamoDB) Transform(ctx context.Context, message *mess.Message) ([]*mess.Message, error) {
	// Skip control messages.
	if message.IsControl() {
		return []*mess.Message{message}, nil
	}

	result := message.Get(mod.conf.Key)
	if !result.IsObject() {
		return nil, fmt.Errorf("transform: proc_aws_dynamodb: key %s: %v", mod.conf.Key, errprocAWSDynamoDBInputNotAnObject)
	}

	// PK is a required field
	pk := json.Get([]byte(result.Raw), "PK").String()
	if pk == "" {
		return nil, fmt.Errorf("transform: proc_aws_dynamodb: key %s: %v", mod.conf.Key, errprocAWSDynamoDBInputMissingPK)
	}

	// SK is an optional field
	sk := json.Get([]byte(result.Raw), "SK").String()

	value, err := mod.dynamodb(ctx, pk, sk)
	if err != nil {
		return nil, fmt.Errorf("transform: proc_aws_dynamodb: %v", err)
	}

	// no match
	if len(value) == 0 {
		return []*mess.Message{message}, nil
	}

	if err := message.Set(mod.conf.SetKey, value); err != nil {
		return nil, fmt.Errorf("transform: proc_aws_dynamodb: %v", err)
	}

	return []*mess.Message{message}, nil
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
