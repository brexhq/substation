//go:build !wasm

package process

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/aws/dynamodb"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/json"
)

var dynamodbAPI dynamodb.API

// errAWSDynamodbInputNotAnObject is returned when the input is not an object.
// Refer to the dynamodb processor documentation for input requirements.
const errAWSDynamodbInputNotAnObject = errors.Error("input is not an object")

// errAWSDynamodbInputMissingPK is returned when the JSON key "PK" is missing in
// the input. Refer to the dynamodb processor documentation for input requirements.
const errAWSDynamodbInputMissingPK = errors.Error("input missing PK")

// awsDynamodb processes data by querying a DynamoDB table and returning all
// matched items as an array of objects. The input must be an object containing
// a partition key ("PK") and optionally containing a sort key ("SK"). This
// processor uses the DynamoDB Query operation, refer to the DynamoDB documentation
// for the Query operation's request syntax and key condition expression patterns:
//
// - https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_Query.html#API_Query_RequestSyntax
//
// - https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Query.html#Query.KeyConditionExpressions
//
// This processor supports the object handling pattern.
type procAWSDynamoDB struct {
	process
	Options procAWSDynamoDBOptions `json:"options"`
}

type procAWSDynamoDBOptions struct {
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

// String returns the processor settings as an object.
func (p procAWSDynamoDB) String() string {
	return toString(p)
}

// Closes resources opened by the processor.
func (p procAWSDynamoDB) Close(context.Context) error {
	return nil
}

// Create a new AWS DynamoDB processor.
func newProcAWSDynamoDB(ctx context.Context, cfg config.Config) (p procAWSDynamoDB, err error) {
	if err = config.Decode(cfg.Settings, &p); err != nil {
		return procAWSDynamoDB{}, err
	}

	p.operator, err = condition.NewOperator(ctx, p.Condition)
	if err != nil {
		return procAWSDynamoDB{}, err
	}

	// fail if required options are missing
	if p.Options.Table == "" || p.Options.KeyConditionExpression == "" {
		return procAWSDynamoDB{}, fmt.Errorf("process: aws_dynamodb: options %+v: %v", p.Options, errors.ErrMissingRequiredOption)
	}

	// only supports JSON, fail; if there are no keys
	if p.Key == "" && p.SetKey == "" {
		return procAWSDynamoDB{}, fmt.Errorf("process: aws_dynamodb: key %s set_key %s: %v", p.Key, p.SetKey, errInvalidDataPattern)
	}
	return p, nil
}

// Batch processes one or more capsules with the processor. Conditions are
// optionally applied to the data to enable processing.
func (p procAWSDynamoDB) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	return batchApply(ctx, capsules, p, p.operator)
}

// Apply processes a capsule with the processor.
func (p procAWSDynamoDB) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	// lazy load API
	if !dynamodbAPI.IsEnabled() {
		dynamodbAPI.Setup()
	}

	result := capsule.Get(p.Key)
	if !result.IsObject() {
		return capsule, fmt.Errorf("process: aws_dynamodb: key %s: %v", p.Key, errAWSDynamodbInputNotAnObject)
	}

	// PK is a required field
	pk := json.Get([]byte(result.Raw), "PK").String()
	if pk == "" {
		return capsule, fmt.Errorf("process: aws_dynamodb: key %s: %v", p.Key, errAWSDynamodbInputMissingPK)
	}

	// SK is an optional field
	sk := json.Get([]byte(result.Raw), "SK").String()

	value, err := p.dynamodb(ctx, pk, sk)
	if err != nil {
		return capsule, fmt.Errorf("process: aws_dynamodb: %v", err)
	}

	// no match
	if len(value) == 0 {
		return capsule, nil
	}

	if err := capsule.Set(p.SetKey, value); err != nil {
		return capsule, fmt.Errorf("process: aws_dynamodb: %v", err)
	}

	return capsule, nil
}

func (p procAWSDynamoDB) dynamodb(ctx context.Context, pk, sk string) ([]map[string]interface{}, error) {
	resp, err := dynamodbAPI.Query(
		ctx,
		p.Options.Table,
		pk, sk,
		p.Options.KeyConditionExpression,
		p.Options.Limit,
		p.Options.ScanIndexForward,
	)
	if err != nil {
		return nil, fmt.Errorf("process: aws_dynamodb: %v", err)
	}

	var items []map[string]interface{}
	for _, i := range resp.Items {
		var item map[string]interface{}
		err = dynamodbattribute.UnmarshalMap(i, &item)
		if err != nil {
			return nil, fmt.Errorf("process: aws_dynamodb: %v", err)
		}

		items = append(items, item)
	}
	return items, nil
}
