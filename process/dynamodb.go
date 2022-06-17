package process

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/internal/aws/dynamodb"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/json"
)

// DynamoDBInvalidSettings is returned when the DynamoDB processor is configured with invalid Input and Output settings.
const DynamoDBInvalidSettings = errors.Error("DynamoDBInvalidSettings")

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

/*
DynamoDB processes data by querying a DynamoDB table and returning all matched items as an array of JSON objects. The processor supports these patterns:
	json:
		{"ddb":{"PK":"foo"}} >>> {"ddb":[{"foo":"bar"}]}

The processor uses this Jsonnet configuration:
	{
		type: 'dynamodb',
		settings: {
			// the input key is expected to be a map containing a partition key ("PK") and an optional sort key ("SK")
			input: {
				key: 'ddb',
			},
			output: {
				key: 'ddb',
			},
			// if the value of the PK is "foo", then this queries DynamoDB by using "foo" as the paritition key value for the table attribute "pk" and returns the last indexed item from the table.
			options: {
				table: 'foo-table',
				key_condition_expression: 'pk = :partitionkeyval',
				limit: 1,
				scan_index_forward: true,
			}
		},
	}
*/
type DynamoDB struct {
	Condition condition.OperatorConfig `json:"condition"`
	Input     string                   `json:"input"`
	Output    string                   `json:"output"`
	Options   DynamoDBOptions          `json:"options"`
}

var dynamodbAPI dynamodb.API

// Slice processes a slice of bytes with the DynamoDB processor. Conditions are optionally applied on the bytes to enable processing.
func (p DynamoDB) Slice(ctx context.Context, s [][]byte) ([][]byte, error) {
	// lazy load API
	if !dynamodbAPI.IsEnabled() {
		dynamodbAPI.Setup()
	}

	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, fmt.Errorf("slicer settings %v: %v", p, err)
	}

	slice := NewSlice(&s)
	for _, data := range s {
		ok, err := op.Operate(data)
		if err != nil {
			return nil, fmt.Errorf("slicer settings %v: %v", p, err)
		}

		if !ok {
			slice = append(slice, data)
			continue
		}

		processed, err := p.Byte(ctx, data)
		if err != nil {
			return nil, fmt.Errorf("slicer: %v", err)
		}
		slice = append(slice, processed)
	}

	return slice, nil
}

// Byte processes bytes with the DynamoDB processor.
func (p DynamoDB) Byte(ctx context.Context, data []byte) ([]byte, error) {
	// lazy load API
	if !dynamodbAPI.IsEnabled() {
		dynamodbAPI.Setup()
	}

	// only supports json, error early if there are no keys
	if p.Input == "" && p.Output == "" {
		return nil, fmt.Errorf("byter settings %v: %v", p, DynamoDBInvalidSettings)
	}

	// json processing
	request := json.Get(data, p.Input)
	if !request.IsObject() {
		return nil, fmt.Errorf("byter settings %v: %v", p, DynamoDBInvalidSettings)
	}

	// PK is a required field
	pk := json.Get([]byte(request.Raw), "PK").String()
	if pk == "" {
		return nil, fmt.Errorf("byter settings %v: %v", p, DynamoDBInvalidSettings)
	}

	sk := json.Get([]byte(request.Raw), "SK").String()

	items, err := p.dynamodb(ctx, pk, sk)
	if err != nil {
		return nil, fmt.Errorf("byter settings %v: %v", p, err)
	}

	// no match
	if len(items) == 0 {
		return data, nil
	}

	return json.Set(data, p.Output, items)
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
		return nil, fmt.Errorf("dynamodb: %v", err)
	}

	var items []map[string]interface{}
	for _, i := range resp.Items {
		var item map[string]interface{}
		err = dynamodbattribute.UnmarshalMap(i, &item)
		if err != nil {
			return nil, fmt.Errorf("dynamodb: %v", err)
		}

		items = append(items, item)
	}
	return items, nil
}
