package transform

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/aws"
	"github.com/brexhq/substation/internal/aws/kinesis"
	_config "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	mess "github.com/brexhq/substation/message"
)

type sendAWSKinesisConfig struct {
	Auth    _config.ConfigAWSAuth `json:"auth"`
	Request _config.ConfigRequest `json:"request"`
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
	if err := _config.Decode(cfg.Settings, &conf); err != nil {
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

func (send *sendAWSKinesis) Transform(ctx context.Context, message *mess.Message) ([]*mess.Message, error) {
	// Lock the transform to prevent concurrent access to the buffer.
	send.mu.Lock()
	defer send.mu.Unlock()

	if message.IsControl() {
		// Flush the buffer.
		for aggregationKey := range send.buffer {
			if send.buffer[aggregationKey].Count == 0 {
				continue
			}

			agg := send.buffer[aggregationKey].Get()
			aggPK := send.buffer[aggregationKey].PartitionKey
			if _, err := send.client.PutRecord(ctx, send.conf.Stream, aggPK, agg); err != nil {
				// PutRecord errors return metadata.
				return nil, fmt.Errorf("send: aws_kinesis: %v", err)
			}
		}

		// Reset the buffer.
		send.buffer = make(map[string]*kinesis.Aggregate)
		return []*mess.Message{message}, nil
	}

	var partitionKey string
	if send.conf.Partition != "" {
		partitionKey = send.conf.Partition
	} else if send.conf.PartitionKey != "" {
		partitionKey = message.Get(send.conf.PartitionKey).String()
	}

	if partitionKey == "" {
		partitionKey = uuid.NewString()
	}

	// Enables redistribution of data across shards by aggregating partition keys into the same payload.
	// This has the intentional side effect where data aggregation is disabled if no partition key is assigned.
	var aggregationKey string
	if send.conf.ShardRedistribution {
		aggregationKey = partitionKey
	}

	if _, ok := send.buffer[aggregationKey]; !ok {
		// Aggregate up to 1MB, the upper limit for Kinesis records.
		send.buffer[aggregationKey] = &kinesis.Aggregate{}
		send.buffer[aggregationKey].New()
	}

	// Add data to the buffer. If the buffer is full, then send the aggregated data.
	if ok := send.buffer[aggregationKey].Add(message.Data(), partitionKey); ok {
		return []*mess.Message{message}, nil
	}

	agg := send.buffer[aggregationKey].Get()
	aggPK := send.buffer[aggregationKey].PartitionKey
	if _, err := send.client.PutRecord(ctx, send.conf.Stream, aggPK, agg); err != nil {
		// PutRecord errors return metadata.
		return nil, fmt.Errorf("send: aws_kinesis: %v", err)
	}

	// Reset the buffer and add the message data.
	send.buffer[aggregationKey].New()
	_ = send.buffer[aggregationKey].Add(message.Data(), partitionKey)

	return []*mess.Message{message}, nil
}
