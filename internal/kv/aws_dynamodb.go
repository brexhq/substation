package kv

import (
	"context"

	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/brexhq/substation/internal/aws/dynamodb"
	"github.com/brexhq/substation/internal/errors"
)

// kvAWSDynamoDB is a read-write key-value store that is backed by an AWS DynamoDB table.
//
// This KV store supports per-item time-to-live (TTL) and has some limitations when
// interacting with DynamoDB:
//
// - Does not support Sort (Range) Keys
//
// - Does not support Global Secondary Indexes
type kvAWSDynamoDB struct {
	// Table is the DynamoDB table that items are read and written to.
	Table      string `json:"table"`
	Attributes struct {
		// PartitionKey is the table attribute where keys are read from and written to.
		PartitionKey string `json:"partition_key"`
		// Value is the table attribute where values are read from and written to.
		Value string `json:"value"`
		// TTL is the table attribute where time-to-live is stored.
		//
		// This option requires the DynamoDB table to be configured with TTL. Learn more
		// about DynamoDB's TTL implementation here: https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/TTL.html.
		TTL string `json:"ttl"`
	} `json:"attributes"`
	api dynamodb.API
}

func (store *kvAWSDynamoDB) String() string {
	return toString(store)
}

// Get retrieves an item from the DynamoDB table. If the item had a time-to-live (TTL)
// configured when it was added and the TTL has passed, then nothing is returned.
//
// This method uses the GetItem API call, which retrieves a single item from the table.
// Learn more about the differences between GetItem and other item retrieval API calls here:
// https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/SQLtoNoSQL.ReadData.html.
func (store *kvAWSDynamoDB) Get(ctx context.Context, key string) (interface{}, error) {
	m := map[string]interface{}{
		store.Attributes.PartitionKey: key,
	}

	resp, err := store.api.GetItem(ctx, store.Table, m)
	if err != nil {
		return "", err
	}

	if val, found := resp.Item[store.Attributes.Value]; found {
		var i interface{}
		if err := dynamodbattribute.Unmarshal(val, &i); err != nil {
			return nil, err
		}

		return i, nil
	}

	return nil, nil
}

// SetWithTTL adds an item to the DynamoDB table.
func (store *kvAWSDynamoDB) Set(ctx context.Context, key string, val interface{}) error {
	m := map[string]interface{}{
		store.Attributes.PartitionKey: key,
		store.Attributes.Value:        val,
	}

	record, err := dynamodbattribute.MarshalMap(m)
	if err != nil {
		return err
	}

	if _, err := store.api.PutItem(ctx, store.Table, record); err != nil {
		return err
	}

	return nil
}

// SetWithTTL adds an item to the DynamoDB table with a time-to-live (TTL) attribute.
func (store *kvAWSDynamoDB) SetWithTTL(ctx context.Context, key string, val interface{}, ttl int64) error {
	if store.Attributes.TTL == "" {
		return errors.ErrMissingRequiredOptions
	}

	m := map[string]interface{}{
		store.Attributes.PartitionKey: key,
		store.Attributes.Value:        val,
		store.Attributes.TTL:          ttl,
	}

	record, err := dynamodbattribute.MarshalMap(m)
	if err != nil {
		return err
	}

	if _, err := store.api.PutItem(ctx, store.Table, record); err != nil {
		return err
	}

	return nil
}

// IsEnabled returns true if the DynamoDB client is ready for use.
func (store *kvAWSDynamoDB) IsEnabled() bool {
	return store.api.IsEnabled()
}

// Setup creates a new DynamoDB client.
func (store *kvAWSDynamoDB) Setup(ctx context.Context) error {
	if store.Table == "" || store.Attributes.PartitionKey == "" {
		return errors.ErrMissingRequiredOptions
	}

	// avoids unnecessary setup
	if store.api.IsEnabled() {
		return nil
	}

	store.api.Setup()

	return nil
}

// Close is unused since connections to DynamoDB are not stateful.
func (store *kvAWSDynamoDB) Close() error {
	return nil
}
