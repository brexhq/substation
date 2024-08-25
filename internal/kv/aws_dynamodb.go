package kv

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/brexhq/substation/v2/config"

	iaws "github.com/brexhq/substation/v2/internal/aws"
	iconfig "github.com/brexhq/substation/v2/internal/config"
)

// kvAWSDynamoDB is a read-write key-value store that is backed by an AWS DynamoDB table.
//
// This KV store supports per-item time-to-live (TTL) and has some limitations when
// interacting with DynamoDB:
//
// - Does not support Global Secondary Indexes
type kvAWSDynamoDB struct {
	AWS        iconfig.AWS `json:"aws"`
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
	client         *dynamodb.Client
}

// Create a new AWS DynamoDB KV store.
func newKVAWSDynamoDB(cfg config.Config) (*kvAWSDynamoDB, error) {
	var store kvAWSDynamoDB
	if err := iconfig.Decode(cfg.Settings, &store); err != nil {
		return nil, err
	}

	if store.AWS.ARN == "" {
		return nil, fmt.Errorf("kv: aws_dynamodb: aws.arn %+v: %v", &store, iconfig.ErrMissingRequiredOption)
	}

	return &store, nil
}

func (store *kvAWSDynamoDB) String() string {
	return toString(store)
}

// Lock adds an item to the DynamoDB table with a conditional check.
func (kv *kvAWSDynamoDB) Lock(ctx context.Context, key string, ttl int64) error {
	attrEx := expression.
		AttributeNotExists(expression.Name(kv.Attributes.PartitionKey)).
		Or(expression.Name(kv.Attributes.TTL).LessThanEqual(expression.Value(time.Now().Unix())))

	m := map[string]interface{}{
		kv.Attributes.PartitionKey: key,
		kv.Attributes.TTL:          ttl,
	}

	if kv.Attributes.SortKey != "" {
		m[kv.Attributes.SortKey] = "substation:kv_store"
	}

	i, err := attributevalue.MarshalMap(m)
	if err != nil {
		return err
	}

	expr, err := expression.NewBuilder().WithCondition(attrEx).Build()
	if err != nil {
		return err
	}

	// If the item already exists and the TTL has not expired, then this returns ErrNoLock. The
	// caller is expected to handle this error and retry the call if necessary.
	input := &dynamodb.PutItemInput{
		TableName:                 aws.String(kv.AWS.ARN),
		Item:                      i,
		ConditionExpression:       expr.Condition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	}

	ctx = context.WithoutCancel(ctx)
	if _, err := kv.client.PutItem(ctx, input); err != nil {
		var ccf *types.ConditionalCheckFailedException
		if errors.As(err, &ccf) {
			return ErrNoLock
		}

		return err
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

	item, err := attributevalue.MarshalMap(m)
	if err != nil {
		return err
	}

	input := &dynamodb.DeleteItemInput{
		TableName: aws.String(store.AWS.ARN),
		Key:       item,
	}

	ctx = context.WithoutCancel(ctx)
	if _, err := store.client.DeleteItem(ctx, input); err != nil {
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

	item, err := attributevalue.MarshalMap(m)
	if err != nil {
		return nil, err
	}

	input := &dynamodb.GetItemInput{
		TableName: aws.String(store.AWS.ARN),
		Key:       item,
	}

	ctx = context.WithoutCancel(ctx)
	resp, err := store.client.GetItem(ctx, input)
	if err != nil {
		return nil, err
	}

	if val, found := resp.Item[store.Attributes.Value]; found {
		var i interface{}
		if err := attributevalue.Unmarshal(val, &i); err != nil {
			return nil, err
		}

		return i, nil
	}

	return nil, nil
}

func (store *kvAWSDynamoDB) Set(ctx context.Context, key string, val interface{}) error {
	m := map[string]interface{}{
		store.Attributes.PartitionKey: key,
		store.Attributes.Value:        val,
	}

	if store.Attributes.SortKey != "" {
		m[store.Attributes.SortKey] = "substation:kv_store"
	}

	item, err := attributevalue.MarshalMap(m)
	if err != nil {
		return err
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String(store.AWS.ARN),
		Item:      item,
	}

	ctx = context.WithoutCancel(ctx)
	if _, err := store.client.PutItem(ctx, input); err != nil {
		return err
	}

	return nil
}

// SetWithTTL adds an item to the DynamoDB table with a time-to-live (TTL) attribute.
func (store *kvAWSDynamoDB) SetWithTTL(ctx context.Context, key string, val interface{}, ttl int64) error {
	if store.Attributes.TTL == "" {
		return iconfig.ErrMissingRequiredOption
	}

	m := map[string]interface{}{
		store.Attributes.PartitionKey: key,
		store.Attributes.Value:        val,
		store.Attributes.TTL:          ttl,
	}

	if store.Attributes.SortKey != "" {
		m[store.Attributes.SortKey] = "substation:kv_store"
	}

	item, err := attributevalue.MarshalMap(m)
	if err != nil {
		return err
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String(store.AWS.ARN),
		Item:      item,
	}

	ctx = context.WithoutCancel(ctx)
	if _, err := store.client.PutItem(ctx, input); err != nil {
		return err
	}

	return nil
}

// SetAddWithTTL adds a value to a set in the DynamoDB table. If the set doesn't exist,
// then a new set is created. If a non-zero TTL is provided, then the TTL attribute is
// updated with the new value.
func (store *kvAWSDynamoDB) SetAddWithTTL(ctx context.Context, key string, val interface{}, ttl int64) error {
	if store.Attributes.Value == "" {
		return iconfig.ErrMissingRequiredOption
	}

	// DynamoDB supports string, number, and binary data types for sets, and
	// numbers are represented as strings.
	var av types.AttributeValue
	switch v := val.(type) {
	case float64:
		av = &types.AttributeValueMemberNS{Value: []string{fmt.Sprintf("%f", v)}}
	case []byte:
		av = &types.AttributeValueMemberBS{Value: [][]byte{v}}
	case string:
		av = &types.AttributeValueMemberSS{Value: []string{v}}
	case []interface{}:
		// The slice of interfaces must be converted depending on the type of all elements.
		// Precedence is given to float64, then []byte, and finally string.
		var ns []string
		var bs [][]byte

		for _, i := range v {
			switch i := i.(type) {
			case float64:
				ns = append(ns, fmt.Sprintf("%f", i))
			case []float64:
				for _, n := range v {
					ns = append(ns, fmt.Sprintf("%f", n))
				}
			case []byte:
				bs = append(bs, i)
			case [][]byte:
				bs = append(bs, i...)
			}
		}

		if len(ns) == len(v) {
			av = &types.AttributeValueMemberNS{Value: ns}

			break
		} else if len(bs) == len(v) {
			av = &types.AttributeValueMemberBS{Value: bs}

			break
		}

		// If the elements are not uniform, then convert all elements to strings.
		var ss []string
		for _, i := range v {
			ss = append(ss, fmt.Sprintf("%v", i))
		}

		av = &types.AttributeValueMemberSS{Value: ss}
	}

	// Overwrite the TTL value if the attribute exists.
	updateEx := expression.Add(expression.Name(store.Attributes.Value), expression.Value(av))
	if store.Attributes.TTL != "" {
		updateEx = updateEx.Set(expression.Name(store.Attributes.TTL), expression.Value(ttl))
	}

	m := map[string]interface{}{
		store.Attributes.PartitionKey: key,
	}

	if store.Attributes.SortKey != "" {
		m[store.Attributes.SortKey] = "substation:kv_store"
	}

	item, err := attributevalue.MarshalMap(m)
	if err != nil {
		return err
	}

	expr, err := expression.NewBuilder().WithUpdate(updateEx).Build()
	if err != nil {
		return err
	}

	input := &dynamodb.UpdateItemInput{
		TableName:                 aws.String(store.AWS.ARN),
		Key:                       item,
		UpdateExpression:          expr.Update(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	}

	ctx = context.WithoutCancel(ctx)
	if _, err := store.client.UpdateItem(ctx, input); err != nil {
		return err
	}

	return nil
}

// IsEnabled returns true if the DynamoDB client is ready for use.
func (store *kvAWSDynamoDB) IsEnabled() bool {
	return store.client != nil
}

// Setup creates a new DynamoDB client.
func (store *kvAWSDynamoDB) Setup(ctx context.Context) error {
	if store.AWS.ARN == "" || store.Attributes.PartitionKey == "" {
		return iconfig.ErrMissingRequiredOption
	}

	// Avoids unnecessary setup.
	if store.client != nil {
		return nil
	}

	awsCfg, err := iaws.New(ctx, iaws.Config{
		Region:  iaws.ParseRegion(store.AWS.ARN),
		RoleARN: store.AWS.AssumeRoleARN,
	})
	if err != nil {
		return err
	}

	store.client = dynamodb.NewFromConfig(awsCfg)

	return nil
}

// Close is unused since connections to DynamoDB are not stateful.
func (store *kvAWSDynamoDB) Close() error {
	return nil
}
