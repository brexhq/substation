package transform

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/aws"
	"github.com/brexhq/substation/internal/aws/kinesis"
	"github.com/brexhq/substation/internal/errors"
	mess "github.com/brexhq/substation/message"
)

type sendAWSKinesisConfig struct {
	Auth    config.ConfigAWSAuth `json:"auth"`
	Request config.ConfigRequest `json:"request"`
	// Stream is the Kinesis Data Stream that records are sent to.
	// TODO(v1.0.0): replace with ARN
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

type sendAWSKinesis struct {
	conf sendAWSKinesisConfig

	// client is safe for concurrent use.
	client kinesis.API

	// buffer is safe for concurrent use.
	mu     sync.Mutex
	buffer map[string]*kinesis.Aggregate
}

func newSendAWSKinesis(_ context.Context, cfg config.Config) (*sendAWSKinesis, error) {
	conf := sendAWSKinesisConfig{}
	if err := config.Decode(cfg.Settings, &conf); err != nil {
		return nil, err
	}

	// Validate required options.
	if conf.Stream == "" {
		return nil, fmt.Errorf("send: aws_kinesis: stream stream: %v", errors.ErrMissingRequiredOption)
	}

	send := sendAWSKinesis{
		conf: conf,
	}

	// Setup the AWS client.
	send.client.Setup(aws.Config{
		Region:     conf.Auth.Region,
		AssumeRole: conf.Auth.AssumeRole,
		MaxRetries: conf.Request.MaxRetries,
	})

	send.mu = sync.Mutex{}
	send.buffer = make(map[string]*kinesis.Aggregate)

	return &send, nil
}

func (*sendAWSKinesis) Close(context.Context) error {
	return nil
}

func (t *sendAWSKinesis) Transform(ctx context.Context, messages ...*mess.Message) ([]*mess.Message, error) {
	// Lock the transform to prevent concurrent access to the buffer.
	t.mu.Lock()
	defer t.mu.Unlock()

	control := false
	for _, message := range messages {
		if message.IsControl() {
			control = true
			continue
		}

		var partitionKey string
		if t.conf.Partition != "" {
			partitionKey = t.conf.Partition
		} else if t.conf.PartitionKey != "" {
			partitionKey = message.Get(t.conf.PartitionKey).String()
		}

		if partitionKey == "" {
			partitionKey = uuid.NewString()
		}

		// Enables redistribution of data across shards by aggregating partition keys into the same payload.
		// This has the intentional side effect where data aggregation is disabled if no partition key is assigned.
		var aggregationKey string
		if t.conf.ShardRedistribution {
			aggregationKey = partitionKey
		}

		if _, ok := t.buffer[aggregationKey]; !ok {
			// Aggregate up to 1MB, the upper limit for Kinesis records.
			t.buffer[aggregationKey] = &kinesis.Aggregate{}
			t.buffer[aggregationKey].New()
		}

		// Add data to the buffer. If the buffer is full, then send the aggregated data.
		ok := t.buffer[aggregationKey].Add(message.Data(), partitionKey)
		if !ok {
			agg := t.buffer[aggregationKey].Get()
			aggPK := t.buffer[aggregationKey].PartitionKey
			if _, err := t.client.PutRecord(ctx, t.conf.Stream, aggPK, agg); err != nil {
				// PutRecord errors return metadata.
				return nil, fmt.Errorf("send: aws_kinesis: %v", err)
			}

			t.buffer[aggregationKey].New()
			t.buffer[aggregationKey].Add(message.Data(), partitionKey)
		}
	}

	// If a control wasn't received, then data stays in the buffer.
	if !control {
		return messages, nil
	}

	// Flush the buffer.
	for aggregationKey := range t.buffer {
		count := t.buffer[aggregationKey].Count

		if count == 0 {
			t.buffer[aggregationKey] = &kinesis.Aggregate{}
			t.buffer[aggregationKey].New()

			continue
		}

		agg := t.buffer[aggregationKey].Get()
		aggPK := t.buffer[aggregationKey].PartitionKey
		if _, err := t.client.PutRecord(ctx, t.conf.Stream, aggPK, agg); err != nil {
			// PutRecord errors return metadata.
			return nil, fmt.Errorf("send: aws_kinesis: %v", err)
		}

		delete(t.buffer, aggregationKey)
	}

	return messages, nil
}
