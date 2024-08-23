package transform

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/internal/aggregate"
	"github.com/brexhq/substation/v2/internal/aws"
	idynamodb "github.com/brexhq/substation/v2/internal/aws/dynamodb"
	iconfig "github.com/brexhq/substation/v2/internal/config"
	"github.com/brexhq/substation/v2/internal/errors"
	"github.com/brexhq/substation/v2/message"
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

	ID     string         `json:"id"`
	Object iconfig.Object `json:"object"`
	Batch  iconfig.Batch  `json:"batch"`
	AWS    iconfig.AWS    `json:"aws"`
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

func newSendAWSDynamoDBPut(_ context.Context, cfg config.Config) (*sendAWSDynamoDBPut, error) {
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

	tf.client.Setup(aws.Config{
		Region:  conf.AWS.Region,
		RoleARN: conf.AWS.RoleARN,
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
				return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
			}

			tf.tforms[i] = t
		}
	}

	return &tf, nil
}

type sendAWSDynamoDBPut struct {
	conf sendAWSDynamoDBConfig

	// client is safe for concurrent use.
	client idynamodb.API

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
