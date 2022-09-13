package sink

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/aws/kinesis"
	"github.com/brexhq/substation/internal/log"
	"github.com/brexhq/substation/internal/metrics"
)

var kinesisAPI kinesis.API

/*
Kinesis sinks data to an AWS Kinesis Data Stream using Kinesis Producer Library (KPL) compliant aggregated records. This sink can automatically redistribute data across shards by retrieving partition keys from JSON data; by default, it uses random strings to avoid hot shards. More information about the KPL and its schema is available here: https://docs.aws.amazon.com/streams/latest/dev/developing-producers-with-kpl.html.

The sink has these settings:
	Stream:
		Kinesis Data Stream that data is sent to
	Partition (optional):
		string that is used as the partition key for the aggregated record
	PartitionKey (optional):
		JSON key-value that is used as the partition key for the aggregated record
	ShardRedistribution (optional):
		determines if data should be redistributed across shards based on the partition key
		if enabled with an empty partition key, then data aggregation is disabled
		defaults to false, data is randomly distributed across shards

When loaded with a factory, the sink uses this JSON configuration:
	{
		"type": "kinesis",
		"settings": {
			"stream": "foo"
		}
	}
*/
type Kinesis struct {
	Stream              string `json:"stream"`
	Partition           string `json:"partition"`
	PartitionKey        string `json:"partition_key"`
	ShardRedistribution bool   `json:"shard_redistribution"`
}

// Send sinks a channel of encapsulated data with the Kinesis sink.
func (sink *Kinesis) Send(ctx context.Context, ch chan config.Capsule, kill chan struct{}) error {
	// matches dimensions for AWS Kinesis Data Streams metrics (https://docs.aws.amazon.com/streams/latest/dev/monitoring-with-cloudwatch.html#kinesis-metricdimensions)
	metricsAttributes := map[string]string{
		"StreamName": sink.Stream,
	}

	if !kinesisAPI.IsEnabled() {
		kinesisAPI.Setup()
	}

	buffer := map[string]*kinesis.Aggregate{}

	for cap := range ch {
		select {
		case <-kill:
			return nil
		default:
			var partitionKey string
			if sink.Partition != "" {
				partitionKey = sink.Partition
			} else if sink.PartitionKey != "" {
				partitionKey = cap.Get(sink.PartitionKey).String()
			}

			if partitionKey == "" {
				partitionKey = uuid.NewString()
			}

			// enables redistribution of data across shards by aggregating partition keys into the same payload
			// this has the intentional side effect where data aggregation is disabled if no partition key is assigned
			var aggregationKey string
			if sink.ShardRedistribution {
				aggregationKey = partitionKey
			}

			if _, ok := buffer[aggregationKey]; !ok {
				// aggregate up to 1MB, the upper limit for Kinesis records
				buffer[aggregationKey] = &kinesis.Aggregate{}
				buffer[aggregationKey].New()
			}

			// add data to the buffer
			// if buffer is full, then send the aggregated data
			ok := buffer[aggregationKey].Add(cap.GetData(), partitionKey)
			if !ok {
				agg := buffer[aggregationKey].Get()
				aggPK := buffer[aggregationKey].PartitionKey
				_, err := kinesisAPI.PutRecord(ctx, agg, sink.Stream, aggPK)
				if err != nil {
					// PutRecord err returns metadata
					return fmt.Errorf("sink kinesis: %v", err)
				}

				log.WithField(
					"stream", sink.Stream,
				).WithField(
					"partition_key", aggPK,
				).WithField(
					"count", buffer[aggregationKey].Count,
				).Debug("put records into Kinesis")

				metrics.Generate(ctx, metrics.Data{
					Attributes: metricsAttributes,
					Name:       "CapsulesSent",
					Value:      buffer[aggregationKey].Count,
				})

				buffer[aggregationKey].New()
				buffer[aggregationKey].Add(cap.GetData(), partitionKey)
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
		_, err := kinesisAPI.PutRecord(ctx, agg, sink.Stream, aggPK)
		if err != nil {
			// PutRecord err returns metadata
			return fmt.Errorf("sink kinesis: %v", err)
		}

		log.WithField(
			"stream", sink.Stream,
		).WithField(
			"partition_key", aggPK,
		).WithField(
			"count", buffer[aggregationKey].Count,
		).Debug("put records into Kinesis")

		metrics.Generate(ctx, metrics.Data{
			Attributes: metricsAttributes,
			Name:       "CapsulesSent",
			Value:      buffer[aggregationKey].Count,
		})
	}

	return nil
}
