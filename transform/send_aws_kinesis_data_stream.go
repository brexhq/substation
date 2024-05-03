package transform

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/aggregate"
	"github.com/brexhq/substation/internal/aws"
	"github.com/brexhq/substation/internal/aws/kinesis"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/message"
	"github.com/google/uuid"
)

// Records greater than 1 MiB in size cannot be
// put into a Kinesis Data Stream.
const sendAWSKinesisDataStreamMessageSizeLimit = 1024 * 1024 * 1

// errSendAWSKinesisDataStreamMessageSizeLimit is returned when data
// exceeds the Kinesis record size limit. If this error occurs, then
// conditions or transforms should be applied to either drop or reduce
// the size of the data.
var errSendAWSKinesisDataStreamMessageSizeLimit = fmt.Errorf("data exceeded size limit")

type sendAWSKinesisDataStreamConfig struct {
	// StreamName is the Kinesis Data Stream that records are sent to.
	StreamName string `json:"stream_name"`
	// UseBatchKeyAsPartitionKey determines if the batch key should be used as the partition key.
	UseBatchKeyAsPartitionKey bool `json:"use_batch_key_as_partition_key"`
	// EnableRecordAggregation determines if records should be aggregated.
	EnableRecordAggregation bool `json:"enable_record_aggregation"`
	// AuxTransforms are applied to batched data before it is sent.
	AuxTransforms []config.Config `json:"auxiliary_transforms"`

	Object iconfig.Object `json:"object"`
	Batch  iconfig.Batch  `json:"batch"`
	AWS    iconfig.AWS    `json:"aws"`
	Retry  iconfig.Retry  `json:"retry"`
}

func (c *sendAWSKinesisDataStreamConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *sendAWSKinesisDataStreamConfig) Validate() error {
	if c.StreamName == "" {
		return fmt.Errorf("stream_name: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

func newSendAWSKinesisDataStream(_ context.Context, cfg config.Config) (*sendAWSKinesisDataStream, error) {
	conf := sendAWSKinesisDataStreamConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: send_aws_kinesis_data_stream: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: send_aws_kinesis_data_stream: %v", err)
	}

	tf := sendAWSKinesisDataStream{
		conf: conf,
	}

	agg, err := aggregate.New(aggregate.Config{
		// Kinesis Data Streams limits batch operations to 500 records and 5MiB.
		Count:    500,
		Size:     sendAWSKinesisDataStreamMessageSizeLimit * 5,
		Duration: conf.Batch.Duration,
	})
	if err != nil {
		return nil, fmt.Errorf("transform: send_aws_kinesis_data_stream: %v", err)
	}
	tf.agg = agg

	if len(conf.AuxTransforms) > 0 {
		tf.tforms = make([]Transformer, len(conf.AuxTransforms))
		for i, c := range conf.AuxTransforms {
			t, err := New(context.Background(), c)
			if err != nil {
				return nil, fmt.Errorf("transform: send_aws_kinesis_data_stream: %v", err)
			}

			tf.tforms[i] = t
		}
	}

	// Setup the AWS client.
	tf.client.Setup(aws.Config{
		Region:          conf.AWS.Region,
		RoleARN:         conf.AWS.RoleARN,
		MaxRetries:      conf.Retry.Count,
		RetryableErrors: conf.Retry.ErrorMessages,
	})

	return &tf, nil
}

type sendAWSKinesisDataStream struct {
	conf sendAWSKinesisDataStreamConfig

	// client is safe for concurrent use.
	client kinesis.API

	mu     sync.Mutex
	agg    *aggregate.Aggregate
	tforms []Transformer
}

func (tf *sendAWSKinesisDataStream) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	tf.mu.Lock()
	defer tf.mu.Unlock()

	if msg.IsControl() {
		for key := range tf.agg.GetAll() {
			if tf.agg.Count(key) == 0 {
				continue
			}

			if err := tf.send(ctx, key); err != nil {
				return nil, fmt.Errorf("transform: send_aws_kinesis_data_stream: %v", err)
			}
		}

		tf.agg.ResetAll()
		return []*message.Message{msg}, nil
	}

	if len(msg.Data()) > sendAWSKinesisDataStreamMessageSizeLimit {
		return nil, fmt.Errorf("transform: send_aws_kinesis_data_stream: %v", errSendAWSKinesisDataStreamMessageSizeLimit)
	}

	// If this value does not exist, then all data is batched together.
	key := msg.GetValue(tf.conf.Object.BatchKey).String()
	if ok := tf.agg.Add(key, msg.Data()); ok {
		return []*message.Message{msg}, nil
	}

	if err := tf.send(ctx, key); err != nil {
		return nil, fmt.Errorf("transform: send_aws_kinesis_data_stream: %v", err)
	}

	// If data cannot be added after reset, then the batch is misconfgured.
	tf.agg.Reset(key)
	if ok := tf.agg.Add(key, msg.Data()); !ok {
		return nil, fmt.Errorf("transform: send_aws_kinesis_data_stream: %v", errSendBatchMisconfigured)
	}

	return []*message.Message{msg}, nil
}

func (tf *sendAWSKinesisDataStream) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}

func (tf *sendAWSKinesisDataStream) send(ctx context.Context, key string) error {
	data, err := withTransforms(ctx, tf.tforms, tf.agg.Get(key))
	if err != nil {
		return err
	}

	var partitionKey string
	switch tf.conf.UseBatchKeyAsPartitionKey {
	case true:
		partitionKey = key
	case false:
		partitionKey = uuid.NewString()
	}

	if tf.conf.EnableRecordAggregation {
		data = tf.aggregateRecords(partitionKey, data)
	}

	if len(data) == 0 {
		return nil
	}

	if _, err := tf.client.PutRecords(ctx, tf.conf.StreamName, partitionKey, data); err != nil {
		return err
	}

	return nil
}

func (tf *sendAWSKinesisDataStream) aggregateRecords(partitionKey string, data [][]byte) [][]byte {
	var records [][]byte

	agg := &kinesis.Aggregate{}
	agg.New()

	for _, d := range data {
		if ok := agg.Add(d, partitionKey); ok {
			continue
		}

		records = append(records, agg.Get())

		agg.New()
		_ = agg.Add(d, partitionKey)
	}

	if agg.Count > 0 {
		records = append(records, agg.Get())
	}

	return records
}
