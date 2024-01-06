package transform

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/aggregate"
	"github.com/brexhq/substation/internal/aws"
	idynamodb "github.com/brexhq/substation/internal/aws/dynamodb"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/message"
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
	// TableName is the DynamoDB table that items are written to.
	TableName string `json:"table_name"`
	// AuxTransforms are applied to batched data before it is sent.
	AuxTransforms []config.Config `json:"auxiliary_transforms"`

	Object iconfig.Object `json:"object"`
	Batch  iconfig.Batch  `json:"batch"`
	AWS    iconfig.AWS    `json:"aws"`
	Retry  iconfig.Retry  `json:"retry"`
}

func (c *sendAWSDynamoDBConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *sendAWSDynamoDBConfig) Validate() error {
	if c.TableName == "" {
		return fmt.Errorf("table_name: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

func newSendAWSDynamoDB(_ context.Context, cfg config.Config) (*sendAWSDynamoDB, error) {
	conf := sendAWSDynamoDBConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: send_aws_dynamodb: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: send_aws_dynamodb: %v", err)
	}

	if conf.Object.SourceKey == "" {
		conf.Object.SourceKey = "@this"
	}

	tf := sendAWSDynamoDB{
		conf: conf,
	}

	tf.client.Setup(aws.Config{
		Region:     conf.AWS.Region,
		RoleARN:    conf.AWS.RoleARN,
		MaxRetries: conf.Retry.Count,
	})

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
				return nil, fmt.Errorf("transform: send_aws_kinesis_data_firehose: %v", err)
			}

			tf.tforms[i] = t
		}
	}

	return &tf, nil
}

type sendAWSDynamoDB struct {
	conf sendAWSDynamoDBConfig

	// client is safe for concurrent use.
	client idynamodb.API

	mu     sync.Mutex
	agg    *aggregate.Aggregate
	tforms []Transformer
}

func (tf *sendAWSDynamoDB) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	tf.mu.Lock()
	defer tf.mu.Unlock()

	if msg.IsControl() {
		for key := range tf.agg.GetAll() {
			if tf.agg.Count(key) == 0 {
				continue
			}

			if err := tf.send(ctx, key); err != nil {
				return nil, fmt.Errorf("transform: send_aws_dynamodb: %v", err)
			}
		}

		tf.agg.ResetAll()
		return []*message.Message{msg}, nil
	}

	if !json.Valid(msg.Data()) {
		return nil, fmt.Errorf("transform: send_aws_dynamodb: %v", errSendAWSDynamoDBNonObject)
	}

	value := msg.GetValue(tf.conf.Object.SourceKey)
	if !value.Exists() {
		return []*message.Message{msg}, nil
	}

	if len(value.Bytes()) > sendAWSDynamoDBItemSizeLimit {
		return nil, fmt.Errorf("transform: send_aws_dynamodb: %v", errSendAWSDynamoDBItemSizeLimit)
	}

	// If this value does not exist, then all data is batched together.
	key := msg.GetValue(tf.conf.Object.BatchKey).String()
	if ok := tf.agg.Add(key, msg.Data()); ok {
		return []*message.Message{msg}, nil
	}

	if err := tf.send(ctx, key); err != nil {
		return nil, fmt.Errorf("transform: send_aws_dynamodb: %v", err)
	}

	// If data cannot be added after reset, then the batch is misconfgured.
	tf.agg.Reset(key)
	if ok := tf.agg.Add(key, msg.Data()); !ok {
		return nil, fmt.Errorf("transform: send_aws_dynamodb: %v", errSendBatchMisconfigured)
	}

	return []*message.Message{msg}, nil
}

func (tf *sendAWSDynamoDB) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}

func (tf *sendAWSDynamoDB) send(ctx context.Context, key string) error {
	data, err := withTransforms(ctx, tf.tforms, tf.agg.Get(key))
	if err != nil {
		return err
	}

	var items []map[string]*dynamodb.AttributeValue
	for _, b := range data {
		m := make(map[string]any)
		for k, v := range bytesToValue(b).Map() {
			m[k] = v.Value()
		}

		i, err := dynamodbattribute.MarshalMap(m)
		if err != nil {
			return err
		}

		items = append(items, i)
	}

	if _, err := tf.client.BatchPutItem(ctx, tf.conf.TableName, items); err != nil {
		return err
	}

	return nil
}
