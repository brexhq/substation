package process

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"

	"github.com/brexhq/substation/config"
	idynamodb "github.com/brexhq/substation/internal/aws/dynamodb"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/json"
)

var dynamodbAPI idynamodb.API

// errdynamodbInputNotAnObject is returned when the input is not a JSON object. Refer to the dynamodb processor documentation for input requirements.
const errdynamodbInputNotAnObject = errors.Error("input is not an object")

// errdynamodbInputMissingPK is returned when the JSON key "PK" is missing in the input. Refer to the dynamodb processor documentation for input requirements.
const errdynamodbInputMissingPK = errors.Error("input missing PK")

/*
dynamodb processes data by querying a dynamodb table and returning all matched items as an array of JSON objects. The input must be a JSON object containing a partition key ("PK") and optionally containing a sort key ("SK"). This processor uses the dynamodb Query operation, refer to the dynamodb documentation for the Query operation's request syntax and key condition expression patterns:

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
type dynamodb struct {
	process
	Options dynamodbOptions `json:"options"`
}

/*
dynamodbOptions contains custom options settings for the dynamodb processor (https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_Query.html#API_Query_RequestSyntax):

	Table:
		dynamodb table to query
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
type dynamodbOptions struct {
	Table                  string `json:"table"`
	KeyConditionExpression string `json:"key_condition_expression"`
	Limit                  int64  `json:"limit"`
	ScanIndexForward       bool   `json:"scan_index_forward"`
}

// Close closes resources opened by the dynamodb processor.
func (p dynamodb) Close(context.Context) error {
	return nil
}

func (p dynamodb) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	capsules, err := conditionalApply(ctx, capsules, p.Condition, p)
	if err != nil {
		return nil, fmt.Errorf("process dynamodb: %v", err)
	}

	return capsules, nil
}

// Apply processes encapsulated data with the dynamodb processor.
func (p dynamodb) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	// error early if required options are missing
	if p.Options.Table == "" || p.Options.KeyConditionExpression == "" {
		return capsule, fmt.Errorf("process dynamodb: options %+v: %v", p.Options, errMissingRequiredOptions)
	}

	// only supports JSON, error early if there are no keys
	if p.Key == "" && p.SetKey == "" {
		return capsule, fmt.Errorf("process dynamodb: inputkey %s outputkey %s: %v", p.Key, p.SetKey, errInvalidDataPattern)
	}

	// lazy load API
	if !dynamodbAPI.IsEnabled() {
		dynamodbAPI.Setup()
	}

	result := capsule.Get(p.Key)
	if !result.IsObject() {
		return capsule, fmt.Errorf("process dynamodb: inputkey %s: %v", p.Key, errdynamodbInputNotAnObject)
	}

	// PK is a required field
	pk := json.Get([]byte(result.Raw), "PK").String()
	if pk == "" {
		return capsule, fmt.Errorf("process dynamodb: inputkey %s: %v", p.Key, errdynamodbInputMissingPK)
	}

	// SK is an optional field
	sk := json.Get([]byte(result.Raw), "SK").String()

	value, err := p.dynamodb(ctx, pk, sk)
	if err != nil {
		return capsule, fmt.Errorf("process dynamodb: %v", err)
	}

	// no match
	if len(value) == 0 {
		return capsule, nil
	}

	if err := capsule.Set(p.SetKey, value); err != nil {
		return capsule, fmt.Errorf("process dynamodb: %v", err)
	}

	return capsule, nil
}

func (p dynamodb) dynamodb(ctx context.Context, pk, sk string) ([]map[string]interface{}, error) {
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
