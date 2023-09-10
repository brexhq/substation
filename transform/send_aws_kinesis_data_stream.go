package transform

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/google/uuid"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/aws"
	"github.com/brexhq/substation/internal/aws/kinesis"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/message"
)

type sendAWSKinesisDataStreamConfig struct {
	AWS   iconfig.AWS   `json:"aws"`
	Retry iconfig.Retry `json:"retry"`

	// Stream is the Kinesis Data Stream that records are sent to.
	Stream string `json:"stream"`
	// Partition is a string that is used as the partition key for each
	// aggregated record.
	//
	// This is optional and defaults to a randomly generated string.
	Partition string `json:"partition"`
	// PartitionKey retrieves a value from an object that sorts records and
	// is used as the partition key for each aggregated record. If used, then
	// this overrides Partition.
	//
	// This is optional and has no default.
	PartitionKey string `json:"partition_key"`
	// ShardRedistribution determines if records should be redistributed
	// across shards based on the partition key.
	//
	// This is optional and defaults to false (data is randomly distributed
	// across shards). If enabled with an empty partition key, then data
	// aggregation is disabled.
	ShardRedistribution bool `json:"shard_redistribution"`
}

func (c *sendAWSKinesisDataStreamConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *sendAWSKinesisDataStreamConfig) Validate() error {
	if c.Stream == "" {
		return fmt.Errorf("stream: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

func newSendAWSKinesisDataStream(_ context.Context, cfg config.Config) (*sendAWSKinesisDataStream, error) {
	conf := sendAWSKinesisDataStreamConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: new_send_aws_kinesis_data_stream: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: new_send_aws_kinesis_data_stream: %v", err)
	}

	tf := sendAWSKinesisDataStream{
		conf: conf,
	}

	// Setup the AWS client.
	tf.client.Setup(aws.Config{
		Region:     conf.AWS.Region,
		AssumeRole: conf.AWS.AssumeRole,
		MaxRetries: conf.Retry.Attempts,
	})

	tf.mu = sync.Mutex{}
	tf.buffer = make(map[string]*kinesis.Aggregate)

	return &tf, nil
}

type sendAWSKinesisDataStream struct {
	conf sendAWSKinesisDataStreamConfig

	// client is safe for concurrent use.
	client kinesis.API

	// buffer is safe for concurrent use.
	mu     sync.Mutex
	buffer map[string]*kinesis.Aggregate
}

func (tf *sendAWSKinesisDataStream) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	// Lock the transform to prevent concurrent access to the buffer.
	tf.mu.Lock()
	defer tf.mu.Unlock()

	if msg.IsControl() {
		// Flush the buffer.
		for aggregationKey := range tf.buffer {
			if tf.buffer[aggregationKey].Count == 0 {
				continue
			}

			agg := tf.buffer[aggregationKey].Get()
			aggPK := tf.buffer[aggregationKey].PartitionKey
			if _, err := tf.client.PutRecord(ctx, tf.conf.Stream, aggPK, agg); err != nil {
				// PutRecord errors return metadata and don't require more information.
				return nil, fmt.Errorf("transform: send_aws_kinesis_data_stream: %v", err)
			}
		}

		// Reset the buffer.
		tf.buffer = make(map[string]*kinesis.Aggregate)
		return []*message.Message{msg}, nil
	}

	var partitionKey string
	if tf.conf.Partition != "" {
		partitionKey = tf.conf.Partition
	} else if tf.conf.PartitionKey != "" {
		partitionKey = msg.GetValue(tf.conf.PartitionKey).String()
	}

	if partitionKey == "" {
		partitionKey = uuid.NewString()
	}

	// Enables redistribution of data across shards by aggregating partition keys into the same payload.
	// This has the intentional side effect where data aggregation is disabled if no partition key is assigned.
	var aggregationKey string
	if tf.conf.ShardRedistribution {
		aggregationKey = partitionKey
	}

	if _, ok := tf.buffer[aggregationKey]; !ok {
		// Aggregate up to 1MB, the upper limit for Kinesis records.
		tf.buffer[aggregationKey] = &kinesis.Aggregate{}
		tf.buffer[aggregationKey].New()
	}

	// Add data to the buffer. If the buffer is full, then send the aggregated data.
	if ok := tf.buffer[aggregationKey].Add(msg.Data(), partitionKey); ok {
		return []*message.Message{msg}, nil
	}

	agg := tf.buffer[aggregationKey].Get()
	aggPK := tf.buffer[aggregationKey].PartitionKey
	if _, err := tf.client.PutRecord(ctx, tf.conf.Stream, aggPK, agg); err != nil {
		// PutRecord errors return metadata and don't require more information.
		return nil, fmt.Errorf("transform: send_aws_kinesis_data_stream: %v", err)
	}

	// Reset the buffer and add the msg data.
	tf.buffer[aggregationKey].New()
	_ = tf.buffer[aggregationKey].Add(msg.Data(), partitionKey)

	return []*message.Message{msg}, nil
}

func (tf *sendAWSKinesisDataStream) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}

func (*sendAWSKinesisDataStream) Close(context.Context) error {
	return nil
}
