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

// errDynamoDBInputNotAnObject is returned when the input is not a JSON object. Refer to the DynamoDB processor documentation for input requirements.
const errDynamoDBInputNotAnObject = errors.Error("input is not an object")

// errDynamoDBInputMissingPK is returned when the JSON key "PK" is missing in the input. Refer to the DynamoDB processor documentation for input requirements.
const errDynamoDBInputMissingPK = errors.Error("input missing PK")

/*
DynamoDB processes data by querying a DynamoDB table and returning all matched items as an array of JSON objects. The input must be a JSON object containing a partition key ("PK") and optionally containing a sort key ("SK"). This processor uses the DynamoDB Query operation, refer to the DynamoDB documentation for the Query operation's request syntax and key condition expression patterns:

- https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_Query.html#API_Query_RequestSyntax

- https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Query.html#Query.KeyConditionExpressions

The processor supports these patterns:
	JSON:
		{"ddb":{"PK":"foo"}} >>> {"ddb":[{"foo":"bar"}]}
		{"ddb":{"PK":"foo","SK":"baz"}} >>> {"ddb":[{"foo":"bar"}]}

When loaded with a factory, the processor uses this JSON configuration:
	{
		"type": "dynamodb",
		"settings": {
			"options": {
				"table": "foo-table",
				"key_condition_expression": "PK = :pk and begins_with(SK, :sk)",
				"limit": 1,
				"scan_index_forward": true
			},
			"input_key": "ddb",
			"output_key": "ddb"
		}
	}
*/
type DynamoDB struct {
	Options   DynamoDBOptions  `json:"options"`
	Condition condition.Config `json:"condition"`
	InputKey  string           `json:"input_key"`
	OutputKey string           `json:"output_key"`
}

/*
DynamoDBOptions contains custom options settings for the DynamoDB processor (https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_Query.html#API_Query_RequestSyntax):
	Table:
		table to query
	KeyConditionExpression:
		key condition expression (see documentation)
	Limit (optional):
		maximum number of items to evaluate
		defaults to evaluating all items
	ScanIndexForward (optional):
		specifies the order of index traversal
		must be one of:
			true (default): traversal is performed in ascending order
			false: traversal is performed in descending order
*/
type DynamoDBOptions struct {
	Table                  string `json:"table"`
	KeyConditionExpression string `json:"key_condition_expression"`
	Limit                  int64  `json:"limit"`
	ScanIndexForward       bool   `json:"scan_index_forward"`
}

// ApplyBatch processes a slice of encapsulated data with the DynamoDB processor. Conditions are optionally applied to the data to enable processing.
func (p DynamoDB) ApplyBatch(ctx context.Context, caps []config.Capsule) ([]config.Capsule, error) {
	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, fmt.Errorf("process dynamodb: %v", err)
	}

	caps, err = conditionallyApplyBatch(ctx, caps, op, p)
	if err != nil {
		return nil, fmt.Errorf("process dynamodb: %v", err)
	}

	return caps, nil
}

// Apply processes encapsulated data with the DynamoDB processor.
func (p DynamoDB) Apply(ctx context.Context, cap config.Capsule) (config.Capsule, error) {
	// error early if required options are missing
	if p.Options.Table == "" || p.Options.KeyConditionExpression == "" {
		return cap, fmt.Errorf("process dynamodb: options %+v: %v", p.Options, errMissingRequiredOptions)
	}

	// only supports JSON, error early if there are no keys
	if p.InputKey == "" && p.OutputKey == "" {
		return cap, fmt.Errorf("process dynamodb: inputkey %s outputkey %s: %v", p.InputKey, p.OutputKey, errInvalidDataPattern)
	}

	// lazy load API
	if !dynamodbAPI.IsEnabled() {
		dynamodbAPI.Setup()
	}

	result := cap.Get(p.InputKey)
	if !result.IsObject() {
		return cap, fmt.Errorf("process dynamodb: inputkey %s: %v", p.InputKey, errDynamoDBInputNotAnObject)
	}

	// PK is a required field
	pk := json.Get([]byte(result.Raw), "PK").String()
	if pk == "" {
		return cap, fmt.Errorf("process dynamodb: inputkey %s: %v", p.InputKey, errDynamoDBInputMissingPK)
	}

	// SK is an optional field
	sk := json.Get([]byte(result.Raw), "SK").String()

	value, err := p.dynamodb(ctx, pk, sk)
	if err != nil {
		return cap, fmt.Errorf("process dynamodb: %v", err)
	}

	// no match
	if len(value) == 0 {
		return cap, nil
	}

	if err := cap.Set(p.OutputKey, value); err != nil {
		return cap, fmt.Errorf("process dynamodb: %v", err)
	}

	return cap, nil
}

func (p DynamoDB) dynamodb(ctx context.Context, pk, sk string) ([]map[string]interface{}, error) {
	resp, err := dynamodbAPI.Query(
		ctx,
		p.Options.Table,
		pk, sk,
		p.Options.KeyConditionExpression,
		p.Options.Limit,
		p.Options.ScanIndexForward,
	)
	if err != nil {
		return nil, fmt.Errorf("process dynamodb: %v", err)
	}

	var items []map[string]interface{}
	for _, i := range resp.Items {
		var item map[string]interface{}
		err = dynamodbattribute.UnmarshalMap(i, &item)
		if err != nil {
			return nil, fmt.Errorf("process dynamodb: %v", err)
		}

		items = append(items, item)
	}
	return items, nil
}
