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

type sendAWSKinesisDataStreamConfig struct {
	Buffer iconfig.Buffer `json:"buffer"`
	AWS    iconfig.AWS    `json:"aws"`
	Retry  iconfig.Retry  `json:"retry"`

	// StreamName is the Kinesis Data Stream that records are sent to.
	StreamName string `json:"stream_name"`
	// Partition is a string that is used as the partition key for each
	// record.
	//
	// This is optional and defaults to a randomly generated string.
	Partition string `json:"partition"`
	// PartitionKey retrieves a value from an object that is used as the
	// partition key for each record. If used, then this overrides Partition.
	//
	// This is optional and has no default.
	PartitionKey string `json:"partition_key"`
	// Aggregation determines if records should be aggregated.
	Aggregation bool `json:"aggregation"`
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

	if conf.PartitionKey != "" {
		conf.Buffer.Key = conf.PartitionKey
	}

	buffer, err := aggregate.New(aggregate.Config{
		// Kinesis Data Streams limits batch operations to 500 records.
		Count: 500,
		// Kinesis Data Streams record size limit is 5MiB.
		Size:     sendAWSKinesisDataStreamMessageSizeLimit * 5,
		Duration: conf.Buffer.Duration,
	})
	if err != nil {
		return nil, fmt.Errorf("transform: send_aws_kinesis_data_stream: %v", err)
	}
	tf.buffer = buffer
	tf.bufferKey = conf.Buffer.Key

	// Setup the AWS client.
	tf.client.Setup(aws.Config{
		Region:     conf.AWS.Region,
		RoleARN:    conf.AWS.RoleARN,
		MaxRetries: conf.Retry.Count,
	})

	return &tf, nil
}

type sendAWSKinesisDataStream struct {
	conf sendAWSKinesisDataStreamConfig

	// client is safe for concurrent use.
	client kinesis.API

	mu        sync.Mutex
	buffer    *aggregate.Aggregate
	bufferKey string
}

func (tf *sendAWSKinesisDataStream) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	tf.mu.Lock()
	defer tf.mu.Unlock()

	if msg.IsControl() {
		for key := range tf.buffer.GetAll() {
			partitionKey := key
			if partitionKey == "" {
				partitionKey = uuid.NewString()
			}

			data := tf.buffer.Get(key)
			switch tf.conf.Aggregation {
			case false:
				if _, err := tf.client.PutRecords(ctx, tf.conf.StreamName, partitionKey, data); err != nil {
					return nil, fmt.Errorf("transform: send_aws_kinesis_data_stream: %v", err)
				}
			case true:
				if err := tf.sendAggregateRecord(ctx, tf.conf.StreamName, partitionKey, data); err != nil {
					return nil, fmt.Errorf("transform: send_aws_kinesis_data_stream: %v", err)
				}
			}
		}

		tf.buffer.ResetAll()
		return []*message.Message{msg}, nil
	}

	if len(msg.Data()) > sendAWSKinesisDataStreamMessageSizeLimit {
		return nil, fmt.Errorf("transform: send_aws_kinesis_data_stream: %v", errSendAWSSNSMessageSizeLimit)
	}

	key := tf.conf.Partition
	if tf.conf.PartitionKey != "" {
		key = msg.GetValue(tf.conf.PartitionKey).String()
	}

	if ok := tf.buffer.Add(key, msg.Data()); ok {
		return []*message.Message{msg}, nil
	}

	partitionKey := key
	if partitionKey == "" {
		partitionKey = uuid.NewString()
	}

	data := tf.buffer.Get(key)
	switch tf.conf.Aggregation {
	case false:
		if _, err := tf.client.PutRecords(ctx, tf.conf.StreamName, partitionKey, data); err != nil {
			return nil, fmt.Errorf("transform: send_aws_kinesis_data_stream: %v", err)
		}
	case true:
		if err := tf.sendAggregateRecord(ctx, tf.conf.StreamName, partitionKey, data); err != nil {
			return nil, fmt.Errorf("transform: send_aws_kinesis_data_stream: %v", err)
		}
	}

	// Reset the buffer and add the msg data.
	tf.buffer.Reset(key)
	_ = tf.buffer.Add(key, msg.Data())

	return []*message.Message{msg}, nil
}

func (tf *sendAWSKinesisDataStream) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}

func (tf *sendAWSKinesisDataStream) sendAggregateRecord(ctx context.Context, stream, partitionKey string, data [][]byte) error {
	agg := &kinesis.Aggregate{}
	agg.New()

	for _, b := range data {
		if ok := agg.Add(b, partitionKey); ok {
			continue
		}

		if _, err := tf.client.PutRecord(ctx, stream, partitionKey, agg.Get()); err != nil {
			return err
		}

		agg.New()
		_ = agg.Add(b, partitionKey)
	}

	if agg.Count > 0 {
		if _, err := tf.client.PutRecord(ctx, stream, partitionKey, agg.Get()); err != nil {
			return err
		}
	}

	return nil
}
