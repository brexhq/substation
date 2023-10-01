package kv

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/aws"
	"github.com/brexhq/substation/internal/aws/dynamodb"
	_config "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
)

// kvAWSDynamoDB is a read-write key-value store that is backed by an AWS DynamoDB table.
//
// This KV store supports per-item time-to-live (TTL) and has some limitations when
// interacting with DynamoDB:
//
// - Does not support Global Secondary Indexes
type kvAWSDynamoDB struct {
	// TableName is the DynamoDB table that items are read and written to.
	TableName  string `json:"table_name"`
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
		// Value is the table attribute where values are read from and written to.
		Value string `json:"value"`
		// TTL is the table attribute where time-to-live is stored.
		//
		// This option requires the DynamoDB table to be configured with TTL. Learn more
		// about DynamoDB's TTL implementation here: https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/TTL.html.
		TTL string `json:"ttl"`
	} `json:"attributes"`
	// ConsistentRead specifies whether or not to use strongly consistent reads.
	//
	// This is optional and defaults to false (eventually consistent reads).
	ConsistentRead bool `json:"consistent_read"`
	api            dynamodb.API
}

// Create a new AWS DynamoDB KV store.
func newKVAWSDyanmoDB(cfg config.Config) (*kvAWSDynamoDB, error) {
	var store kvAWSDynamoDB
	if err := _config.Decode(cfg.Settings, &store); err != nil {
		return nil, err
	}

	if store.TableName == "" {
		return nil, fmt.Errorf("kv: aws_dynamodb: table %+v: %v", &store, errors.ErrMissingRequiredOption)
	}

	return &store, nil
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

	if store.Attributes.SortKey != "" {
		m[store.Attributes.SortKey] = "substation:kv_store"
	}

	resp, err := store.api.GetItem(ctx, store.TableName, m, store.ConsistentRead)
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

	if store.Attributes.SortKey != "" {
		m[store.Attributes.SortKey] = "substation:kv_store"
	}

	record, err := dynamodbattribute.MarshalMap(m)
	if err != nil {
		return err
	}

	if _, err := store.api.PutItem(ctx, store.TableName, record); err != nil {
		return err
	}

	return nil
}

// SetWithTTL adds an item to the DynamoDB table with a time-to-live (TTL) attribute.
func (store *kvAWSDynamoDB) SetWithTTL(ctx context.Context, key string, val interface{}, ttl int64) error {
	if store.Attributes.TTL == "" {
		return errors.ErrMissingRequiredOption
	}

	m := map[string]interface{}{
		store.Attributes.PartitionKey: key,
		store.Attributes.Value:        val,
		store.Attributes.TTL:          ttl,
	}

	if store.Attributes.SortKey != "" {
		m[store.Attributes.SortKey] = "substation:kv_store"
	}

	record, err := dynamodbattribute.MarshalMap(m)
	if err != nil {
		return err
	}

	if _, err := store.api.PutItem(ctx, store.TableName, record); err != nil {
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
	if store.TableName == "" || store.Attributes.PartitionKey == "" {
		return errors.ErrMissingRequiredOption
	}

	// Avoids unnecessary setup.
	if store.api.IsEnabled() {
		return nil
	}

	store.api.Setup(aws.Config{})

	return nil
}

// Close is unused since connections to DynamoDB are not stateful.
func (store *kvAWSDynamoDB) Close() error {
	return nil
}
