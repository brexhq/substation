package kv

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/aws"
	"github.com/brexhq/substation/internal/aws/dynamodb"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
)

// kvAWSDynamoDB is a read-write key-value store that is backed by an AWS DynamoDB table.
//
// This KV store supports per-item time-to-live (TTL) and has some limitations when
// interacting with DynamoDB:
//
// - Does not support Global Secondary Indexes
type kvAWSDynamoDB struct {
	AWS   iconfig.AWS   `json:"aws"`
	Retry iconfig.Retry `json:"retry"`
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
	client         dynamodb.API
}

// Create a new AWS DynamoDB KV store.
func newKVAWSDynamoDB(cfg config.Config) (*kvAWSDynamoDB, error) {
	var store kvAWSDynamoDB
	if err := iconfig.Decode(cfg.Settings, &store); err != nil {
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

// Lock adds an item to the DynamoDB table with a conditional check.
func (kv *kvAWSDynamoDB) Lock(ctx context.Context, key string, ttl int64) error {
	attr := map[string]interface{}{
		kv.Attributes.PartitionKey: key,
		kv.Attributes.TTL:          ttl,
	}

	if kv.Attributes.SortKey != "" {
		attr[kv.Attributes.SortKey] = "substation:kv_store"
	}

	// Since the sort key is optional and static, it is not included in the check.
	exp := "attribute_not_exists(#pk) OR #ttl <= :now"
	expAttrNames := map[string]*string{
		"#pk":  &kv.Attributes.PartitionKey,
		"#ttl": &kv.Attributes.TTL,
	}
	expAttrVals := map[string]interface{}{
		":now": time.Now().Unix(),
	}

	a, err := dynamodbattribute.MarshalMap(attr)
	if err != nil {
		return err
	}

	v, err := dynamodbattribute.MarshalMap(expAttrVals)
	if err != nil {
		return err
	}

	// If the item already exists and the TTL has not expired, then this returns ErrNoLock. The
	// caller is expected to handle this error and retry the call if necessary.
	if _, err := kv.client.PutItemWithCondition(ctx, kv.TableName, a, exp, expAttrNames, v); err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "ConditionalCheckFailedException" {
				return ErrNoLock
			}
		} else {
			return err
		}
	}

	return nil
}

func (store *kvAWSDynamoDB) Unlock(ctx context.Context, key string) error {
	m := map[string]interface{}{
		store.Attributes.PartitionKey: key,
	}

	if store.Attributes.SortKey != "" {
		m[store.Attributes.SortKey] = "substation:kv_store"
	}

	item, err := dynamodbattribute.MarshalMap(m)
	if err != nil {
		return err
	}

	if _, err := store.client.DeleteItem(ctx, store.TableName, item); err != nil {
		return err
	}

	return nil
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

	resp, err := store.client.GetItem(ctx, store.TableName, m, store.ConsistentRead)
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

	if _, err := store.client.PutItem(ctx, store.TableName, record); err != nil {
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

	if _, err := store.client.PutItem(ctx, store.TableName, record); err != nil {
		return err
	}

	return nil
}

// IsEnabled returns true if the DynamoDB client is ready for use.
func (store *kvAWSDynamoDB) IsEnabled() bool {
	return store.client.IsEnabled()
}

// Setup creates a new DynamoDB client.
func (store *kvAWSDynamoDB) Setup(ctx context.Context) error {
	if store.TableName == "" || store.Attributes.PartitionKey == "" {
		return errors.ErrMissingRequiredOption
	}

	// Avoids unnecessary setup.
	if store.client.IsEnabled() {
		return nil
	}

	store.client.Setup(aws.Config{
		Region:          store.AWS.Region,
		RoleARN:         store.AWS.RoleARN,
		MaxRetries:      store.Retry.Count,
		RetryableErrors: store.Retry.ErrorMessages,
	})

	return nil
}

// Close is unused since connections to DynamoDB are not stateful.
func (store *kvAWSDynamoDB) Close() error {
	return nil
}
