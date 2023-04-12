package sink

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/aws/kinesis"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/log"
)

var kinesisAPI kinesis.API

// awsKinesis sinks data to an AWS Kinesis Data Stream using Kinesis Producer Library (KPL) compliant aggregated records.
//
// More information about the KPL and its schema is available here: https://docs.aws.amazon.com/streams/latest/dev/developing-producers-with-kpl.html.
type sinkAWSKinesis struct {
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

// Create a new AWS Kinesis sink.
func newSinkAWSKinesis(cfg config.Config) (s sinkAWSKinesis, err error) {
	if err = config.Decode(cfg.Settings, &s); err != nil {
		return sinkAWSKinesis{}, err
	}

	if s.Stream == "" {
		return sinkAWSKinesis{}, fmt.Errorf("sink: aws_kinesis: stream stream: %v", errors.ErrMissingRequiredOption)
	}

	return s, nil
}

// Send sinks a channel of encapsulated data with the sink.
func (s sinkAWSKinesis) Send(ctx context.Context, ch *config.Channel) error {
	if !kinesisAPI.IsEnabled() {
		kinesisAPI.Setup()
	}

	buffer := map[string]*kinesis.Aggregate{}

	for capsule := range ch.C {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			var partitionKey string
			if s.Partition != "" {
				partitionKey = s.Partition
			} else if s.PartitionKey != "" {
				partitionKey = capsule.Get(s.PartitionKey).String()
			}

			if partitionKey == "" {
				partitionKey = uuid.NewString()
			}

			// enables redistribution of data across shards by aggregating partition keys into the same payload
			// this has the intentional side effect where data aggregation is disabled if no partition key is assigned
			var aggregationKey string
			if s.ShardRedistribution {
				aggregationKey = partitionKey
			}

			if _, ok := buffer[aggregationKey]; !ok {
				// aggregate up to 1MB, the upper limit for Kinesis records
				buffer[aggregationKey] = &kinesis.Aggregate{}
				buffer[aggregationKey].New()
			}

			// add data to the buffer
			// if buffer is full, then send the aggregated data
			ok := buffer[aggregationKey].Add(capsule.Data(), partitionKey)
			if !ok {
				agg := buffer[aggregationKey].Get()
				aggPK := buffer[aggregationKey].PartitionKey
				_, err := kinesisAPI.PutRecord(ctx, agg, s.Stream, aggPK)
				if err != nil {
					// PutRecord err returns metadata
					return fmt.Errorf("sink: aws_kinesis: %v", err)
				}

				log.WithField(
					"stream", s.Stream,
				).WithField(
					"partition_key", aggPK,
				).WithField(
					"count", buffer[aggregationKey].Count,
				).Debug("put records into Kinesis")

				buffer[aggregationKey].New()
				buffer[aggregationKey].Add(capsule.Data(), partitionKey)
			}
		}
	}

	// iterate and send remaining buffers
	for aggregationKey := range buffer {
		count := buffer[aggregationKey].Count

		if count == 0 {
			continue
		}

		agg := buffer[aggregationKey].Get()
		aggPK := buffer[aggregationKey].PartitionKey
		_, err := kinesisAPI.PutRecord(ctx, agg, s.Stream, aggPK)
		if err != nil {
			// PutRecord err returns metadata
			return fmt.Errorf("sink: aws_kinesis: %v", err)
		}

		log.WithField(
			"stream", s.Stream,
		).WithField(
			"partition_key", aggPK,
		).WithField(
			"count", buffer[aggregationKey].Count,
		).Debug("put records into Kinesis")
	}

	return nil
}
