package process

import (
	"context"

	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/internal/aws/dynamodb"
	"github.com/brexhq/substation/internal/json"
)

/*
DynamoDBInput contains custom input settings for this processor.

PartitionKey: the JSON key that is used as the paritition key value for the DynamoDB query
SortKey (optional): the JSON key that is used as the sort /range key value for the DynamoDB query
*/
type DynamoDBInput struct {
	PartitionKey string `mapstructure:"partition_key"`
	SortKey      string `mapstructure:"sort_key"`
}

/*
DynamoDBOptions contain custom options settings for this processor. Refer to the DynamoDB API query documentation for more information: https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_Query.html.

A common use for this processor is to return the most recent item from a DynamoDB table based on partition and sort keys. This can be achieved by setting Limit to 1 and ScanIndexForward to false.

Table: the DynamoDB table to query
KeyConditionExpression: the DynamoDB key condition expression
Limit (optional): the number of result items to return
ScanIndexForward (optional): reverses the order of item results
*/
type DynamoDBOptions struct {
	Table                  string `mapstructure:"table"`
	KeyConditionExpression string `mapstructure:"key_condition_expression"`
	Limit                  int64  `mapstructure:"limit"`
	ScanIndexForward       bool   `mapstructure:"scan_index_forward"`
}

// DynamoDB implements the Byter and Channeler interfaces and queries DynamoDB and returns all matched items as an array of JSON objects. More information is available in the README.
type DynamoDB struct {
	Condition condition.OperatorConfig `mapstructure:"condition"`
	Input     DynamoDBInput            `mapstructure:"input"`
	Output    Output                   `mapstructure:"output"`
	Options   DynamoDBOptions          `mapstructure:"options"`
}

var dynamodbAPI dynamodb.API

// Channel processes a data channel of bytes with this processor. Conditions can be optionally applied on the channel data to enable processing.
func (p DynamoDB) Channel(ctx context.Context, ch <-chan []byte) (<-chan []byte, error) {
	var array [][]byte

	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, err
	}

	for data := range ch {
		ok, err := op.Operate(data)
		if err != nil {
			return nil, err
		}

		if !ok {
			array = append(array, data)
			continue
		}

		processed, err := p.Byte(ctx, data)
		if err != nil {
			return nil, err
		}
		array = append(array, processed)
	}

	output := make(chan []byte, len(array))
	for _, x := range array {
		output <- x
	}
	close(output)
	return output, nil

}

// Byte processes a byte slice with this processor
func (p DynamoDB) Byte(ctx context.Context, data []byte) ([]byte, error) {
	if !dynamodbAPI.IsEnabled() {
		dynamodbAPI.Setup()
	}

	pk := json.Get(data, p.Input.PartitionKey)
	if pk.Type.String() == "Null" {
		return data, nil
	}

	sk := json.Get(data, p.Input.SortKey)
	if !pk.IsArray() && !sk.IsArray() {
		items, err := p.dynamodb(ctx, pk.String(), sk.String())
		if err != nil {
			return nil, err
		}

		// no match
		if len(items) == 0 {
			return data, nil
		}

		return json.Set(data, p.Output.Key, items)
	}

	var array []interface{}
	for i := 0; i < len(pk.Array()); i++ {
		pks := pk.Array()[i]
		sks := sk.Array()[i]

		items, err := p.dynamodb(ctx, pks.String(), sks.String())
		if err != nil {
			return nil, err
		}
		array = append(array, items)
	}

	if len(array) == 0 {
		return data, nil
	}

	return json.Set(data, p.Output.Key, array)
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
