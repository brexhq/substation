package sink

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/internal/aws/kinesis"
	"github.com/brexhq/substation/internal/json"
	"github.com/brexhq/substation/internal/log"
)

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

The sink uses this Jsonnet configuration:
	{
		type: 'kinesis',
		settings: {
			stream: 'foo',
			partition: 'bar',
		},
	}
*/
type Kinesis struct {
	Stream              string `json:"stream"`
	Partition           string `json:"partition"`
	PartitionKey        string `json:"partition_key"`
	ShardRedistribution bool   `json:"shard_redistribution"`
}

var kinesisAPI kinesis.API

// Send sinks a channel of bytes with the Kinesis sink.
func (sink *Kinesis) Send(ctx context.Context, ch chan []byte, kill chan struct{}) error {
	if !kinesisAPI.IsEnabled() {
		kinesisAPI.Setup()
	}

	buffer := map[string]*kinesis.Aggregate{}

	for data := range ch {
		select {
		case <-kill:
			return nil
		default:
			var partitionKey string
			if sink.Partition != "" {
				partitionKey = sink.Partition
			} else if sink.PartitionKey != "" {
				partitionKey = json.Get(data, sink.PartitionKey).String()
			}

			if partitionKey == "" {
				partitionKey = randomString()
			}

			// enables redistribution of data across shards by aggregating partition keys into the same payload
			// this has the intentional side effect where data aggregation is disabled if no partition key is assigned
			var aggregationKey string
			if sink.ShardRedistribution {
				aggregationKey = partitionKey
			}

			if _, ok := buffer[aggregationKey]; !ok {
				buffer[aggregationKey] = &kinesis.Aggregate{}
				buffer[aggregationKey].New()
			}

			ok := buffer[aggregationKey].Add(data, partitionKey)
			if !ok {
				agg := buffer[aggregationKey].Get()
				aggPK := buffer[aggregationKey].PartitionKey
				_, err := kinesisAPI.PutRecord(ctx, agg, sink.Stream, aggPK)
				if err != nil {
					return fmt.Errorf("err failed to put records into Kinesis stream %s: %v", sink.Stream, err)
				}

				log.WithField(
					"count", buffer[aggregationKey].Count,
				).WithField(
					"partition_key", aggPK,
				).Debug("put records into Kinesis")

				buffer[aggregationKey].New()
				buffer[aggregationKey].Add(data, partitionKey)
			}
		}
	}

	for aggregationKey := range buffer {
		count := buffer[aggregationKey].Count

		if count == 0 {
			continue
		}

		agg := buffer[aggregationKey].Get()
		aggPK := buffer[aggregationKey].PartitionKey
		_, err := kinesisAPI.PutRecord(ctx, agg, sink.Stream, aggPK)
		if err != nil {
			return fmt.Errorf("err failed to put records into Kinesis stream %s: %v", sink.Stream, err)
		}

		log.WithField(
			"count", buffer[aggregationKey].Count,
		).WithField(
			"partition_key", aggPK,
		).Debug("put records into Kinesis")
	}

	return nil
}
