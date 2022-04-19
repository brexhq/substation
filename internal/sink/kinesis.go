package sink

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/internal/aws/kinesis"
	"github.com/brexhq/substation/internal/json"
	"github.com/brexhq/substation/internal/log"
)

/*
Kinesis implements the Sink interface and puts aggregated records into a Kinesis Data Stream. More information is available in the README.

Stream: the Kinesis Data Stream to put data into
PartitionKey (optional): uses the value from this JSON key as the stream partition key; defaults to a random string to avoid hot shards
*/
type Kinesis struct {
	api          kinesis.API
	Stream       string `mapstructure:"stream"`
	PartitionKey string `mapstructure:"partition_key"`
}

// Send sends a channel of bytes to the Kinesis Data Stream stream defined by this sink.
func (sink *Kinesis) Send(ctx context.Context, ch chan []byte, kill chan struct{}) error {
	if !sink.api.IsEnabled() {
		sink.api.Setup()
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

				_, err := sink.api.PutRecord(ctx, aggData, sink.Stream, aggPk)
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

		_, err := sink.api.PutRecord(ctx, aggData, sink.Stream, aggPk)
		if err != nil {
			return fmt.Errorf("err failed to put records into Kinesis stream %s: %v", sink.Stream, err)
		}

		log.WithField(
			"count", agg.Count,
		).Debug("put records into Kinesis")
	}

	return nil
}
