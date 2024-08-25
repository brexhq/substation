package transform

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"

	"github.com/brexhq/substation/v2/internal/aggregate"
	"github.com/brexhq/substation/v2/internal/aws"
	iconfig "github.com/brexhq/substation/v2/internal/config"
)

// Items greater than 400 KB in size cannot be put into DynamoDB.
const sendAWSDynamoDBItemSizeLimit = 1024 * 400

// errSendAWSDynamoDBItemSizeLimit is returned when data exceeds the
// DynamoDB item size limit. If this error occurs, then drop or reduce
// the size of the data before attempting to write it to DynamoDB.
var errSendAWSDynamoDBItemSizeLimit = fmt.Errorf("data exceeded size limit")

// errSendAWSDynamoDBNonObject is returned when non-object data is sent to the transform.
//
// If this error occurs, then parse the data into an object (or drop invalid objects)
// before attempting to send the data.
var errSendAWSDynamoDBNonObject = fmt.Errorf("input must be object")

type sendAWSDynamoDBConfig struct {
	// AuxTransforms are applied to batched data before it is sent.
	AuxTransforms []config.Config `json:"auxiliary_transforms"`

	ID     string         `json:"id"`
	Object iconfig.Object `json:"object"`
	AWS    iconfig.AWS    `json:"aws"`
	Batch  iconfig.Batch  `json:"batch"`
}

func (c *sendAWSDynamoDBConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *sendAWSDynamoDBConfig) Validate() error {
	if c.AWS.ARN == "" {
		return fmt.Errorf("aws.arn: %v", iconfig.ErrMissingRequiredOption)
	}

	return nil
}

func newSendAWSDynamoDBPut(ctx context.Context, cfg config.Config) (*sendAWSDynamoDBPut, error) {
	conf := sendAWSDynamoDBConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform send_aws_dynamodb_put: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "send_aws_dynamodb_put"
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	tf := sendAWSDynamoDBPut{
		conf: conf,
	}

	awsCfg, err := aws.New(ctx, aws.Config{
		Region:  aws.ParseRegion(conf.AWS.ARN),
		RoleARN: conf.AWS.AssumeRoleARN,
	})
	if err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	tf.client = dynamodb.NewFromConfig(awsCfg)

	agg, err := aggregate.New(aggregate.Config{
		// DynamoDB limits batch operations to 25 records and 16 MiB.
		Count:    25,
		Size:     1000 * 1000 * 16,
		Duration: conf.Batch.Duration,
	})
	if err != nil {
		return nil, err
	}
	tf.agg = agg

	if len(conf.AuxTransforms) > 0 {
		tf.tforms = make([]Transformer, len(conf.AuxTransforms))
		for i, c := range conf.AuxTransforms {
			t, err := New(context.Background(), c)
			if err != nil {
				return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
			}

			tf.tforms[i] = t
		}
	}

	return &tf, nil
}

type sendAWSDynamoDBPut struct {
	conf   sendAWSDynamoDBConfig
	client *dynamodb.Client

	mu     sync.Mutex
	agg    *aggregate.Aggregate
	tforms []Transformer
}

func (tf *sendAWSDynamoDBPut) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	tf.mu.Lock()
	defer tf.mu.Unlock()

	if msg.IsControl() {
		for key := range tf.agg.GetAll() {
			if tf.agg.Count(key) == 0 {
				continue
			}

			if err := tf.send(ctx, key); err != nil {
				return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
			}
		}

		tf.agg.ResetAll()
		return []*message.Message{msg}, nil
	}

	if !json.Valid(msg.Data()) {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, errSendAWSDynamoDBNonObject)
	}

	if len(msg.Data()) > sendAWSDynamoDBItemSizeLimit {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, errSendAWSDynamoDBItemSizeLimit)
	}

	// If this value does not exist, then all data is batched together.
	key := msg.GetValue(tf.conf.Object.BatchKey).String()
	if ok := tf.agg.Add(key, msg.Data()); ok {
		return []*message.Message{msg}, nil
	}

	if err := tf.send(ctx, key); err != nil {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
	}

	// If data cannot be added after reset, then the batch is misconfgured.
	tf.agg.Reset(key)
	if ok := tf.agg.Add(key, msg.Data()); !ok {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, errSendBatchMisconfigured)
	}

	return []*message.Message{msg}, nil
}

func (tf *sendAWSDynamoDBPut) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}

func (tf *sendAWSDynamoDBPut) send(ctx context.Context, key string) error {
	data, err := withTransforms(ctx, tf.tforms, tf.agg.Get(key))
	if err != nil {
		return err
	}

	var attrs []map[string]types.AttributeValue
	for _, b := range data {
		var item map[string]interface{}
		if err := json.Unmarshal(b, &item); err != nil {
			return fmt.Errorf("transform %s: %v", tf.conf.ID, err)
		}

		m, err := attributevalue.MarshalMap(item)
		if err != nil {
			return fmt.Errorf("transform %s: %v", tf.conf.ID, err)
		}

		attrs = append(attrs, m)
	}

	ctx = context.WithoutCancel(ctx)
	return tf.putItems(ctx, attrs)
}

func (tf *sendAWSDynamoDBPut) putItems(ctx context.Context, attrs []map[string]types.AttributeValue) error {
	var items []types.WriteRequest
	for _, attr := range attrs {
		items = append(items, types.WriteRequest{
			PutRequest: &types.PutRequest{
				Item: attr,
			},
		})
	}

	input := &dynamodb.BatchWriteItemInput{
		RequestItems: map[string][]types.WriteRequest{
			tf.conf.AWS.ARN: items,
		},
	}

	resp, err := tf.client.BatchWriteItem(ctx, input)
	if err != nil {
		var e *types.ProvisionedThroughputExceededException
		if errors.As(err, &e) {
			var retry []map[string]types.AttributeValue

			for _, item := range resp.UnprocessedItems[tf.conf.AWS.ARN] {
				retry = append(retry, item.PutRequest.Item)
			}

			if len(retry) > 0 {
				return tf.putItems(ctx, retry)
			}
		} else {
			return fmt.Errorf("transform %s: %v", tf.conf.ID, err)
		}
	}

	return nil
}
