package sink

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/internal/aws/kinesis"
	"github.com/brexhq/substation/internal/json"
	"github.com/brexhq/substation/internal/log"
)

/*
Kinesis sinks data to an AWS Kinesis Data Stream using Kinesis Producer Library (KPL) compliant aggregated records. More information about the KPL and its schema is available here: https://docs.aws.amazon.com/streams/latest/dev/developing-producers-with-kpl.html.

The sink has these settings:
	Stream:
		Kinesis Data Stream that data is sent to
	PartitionKey (optional):
		JSON key-value that is used as the partition key for the aggregated record
		defaults to a random string to avoid hot shards

The sink uses this Jsonnet configuration:
	{
		type: 'kinesis',
		settings: {
			stream: 'foo',
		},
	}
*/
type Kinesis struct {
	Stream       string `json:"stream"`
	PartitionKey string `json:"partition_key"`
}

var kinesisAPI kinesis.API

// Send sinks a channel of bytes with the Kinesis sink.
func (sink *Kinesis) Send(ctx context.Context, ch chan []byte, kill chan struct{}) error {
	if !kinesisAPI.IsEnabled() {
		kinesisAPI.Setup()
	}

	agg := kinesis.Aggregate{}
	agg.New()

	for data := range ch {
		select {
		case <-kill:
			return nil
		default:
			var pk string
			if sink.PartitionKey == "" {
				pk = randomString()
			} else if json.Valid(data) {
				pk = json.Get(data, sink.PartitionKey).String()
			} else {
				// can only parse a partition key from JSON
				// if this error occurs, then convert the data to JSON
				// 	or remove the PartitionKey configuration
				return fmt.Errorf("err failed to parse partition key %s due to invalid JSON: %v", sink.PartitionKey, json.JSONInvalidData)
			}

			ok := agg.Add(data, pk)
			if !ok {
				aggData := agg.Get()
				aggPk := agg.PartitionKey

				_, err := kinesisAPI.PutRecord(ctx, aggData, sink.Stream, aggPk)
				if err != nil {
					return fmt.Errorf("err failed to put records into Kinesis stream %s: %v", sink.Stream, err)
				}

				log.WithField(
					"count", agg.Count,
				).Debug("put records into Kinesis")

				agg.New()
				agg.Add(data, pk)
			}
		}
	}

	if agg.Count > 0 {
		aggData := agg.Get()
		aggPk := agg.PartitionKey

		_, err := kinesisAPI.PutRecord(ctx, aggData, sink.Stream, aggPk)
		if err != nil {
			return fmt.Errorf("err failed to put records into Kinesis stream %s: %v", sink.Stream, err)
		}

		log.WithField(
			"count", agg.Count,
		).Debug("put records into Kinesis")
	}

	return nil
}
