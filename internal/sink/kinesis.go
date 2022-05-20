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
	PartitionKey (optional):
		JSON key-value that is used as the partition key for the aggregated record
		defaults to a random string to avoid hot shards, which also disables record aggregation

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
	Stream       string `json:"stream"`
	Partition    string `json:"partition"`
	PartitionKey string `json:"partition_key"`
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
			partitionKey := randomString()
			if sink.PartitionKey != "" {
				pk := json.Get(data, sink.PartitionKey).String()
				if pk != "" {
					partitionKey = pk
				}
			}

			if _, ok := buffer[partitionKey]; !ok {
				buffer[partitionKey] = &kinesis.Aggregate{}
				buffer[partitionKey].New()
			}

			ok := buffer[partitionKey].Add(data, partitionKey)
			if !ok {
				agg := buffer[partitionKey].Get()
				_, err := kinesisAPI.PutRecord(ctx, agg, sink.Stream, partitionKey)
				if err != nil {
					return fmt.Errorf("err failed to put records into Kinesis stream %s: %v", sink.Stream, err)
				}

				log.WithField(
					"count", buffer[partitionKey].Count,
				).Debug("put records into Kinesis")

				buffer[partitionKey].New()
				buffer[partitionKey].Add(data, partitionKey)
			}
		}
	}

	for partitionKey := range buffer {
		count := buffer[partitionKey].Count

		if count == 0 {
			continue
		}

		agg := buffer[partitionKey].Get()
		_, err := kinesisAPI.PutRecord(ctx, agg, sink.Stream, partitionKey)
		if err != nil {
			return fmt.Errorf("err failed to put records into Kinesis stream %s: %v", sink.Stream, err)
		}

		log.WithField(
			"count", buffer[partitionKey].Count,
		).Debug("put records into Kinesis")
	}

	return nil
}
