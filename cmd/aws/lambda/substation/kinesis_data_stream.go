package main

import (
	"context"
	"encoding/json"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/service/kinesis/types"
	"github.com/awslabs/kinesis-aggregation/go/v2/deaggregator"
	"golang.org/x/sync/errgroup"

	"github.com/brexhq/substation/v2"
	"github.com/brexhq/substation/v2/message"

	"github.com/brexhq/substation/v2/internal/channel"
)

type kinesisStreamMetadata struct {
	ApproximateArrivalTimestamp time.Time `json:"approximateArrivalTimestamp"`
	Stream                      string    `json:"stream"`
	PartitionKey                string    `json:"partitionKey"`
	SequenceNumber              string    `json:"sequenceNumber"`
}

func kinesisStreamHandler(ctx context.Context, event events.KinesisEvent) error {
	// Retrieve and load configuration.
	conf, err := getConfig(ctx)
	if err != nil {
		return err
	}

	cfg := customConfig{}
	if err := json.NewDecoder(conf).Decode(&cfg); err != nil {
		return err
	}

	sub, err := substation.New(ctx, cfg.Config)
	if err != nil {
		return err
	}

	ch := channel.New[*message.Message]()
	group, ctx := errgroup.WithContext(ctx)

	// Data transformation. Transforms are executed concurrently using a worker pool
	// managed by an errgroup. Each message is processed in a separate goroutine.
	group.Go(func() error {
		tfGroup, tfCtx := errgroup.WithContext(ctx)
		tfGroup.SetLimit(cfg.Concurrency)

		for message := range ch.Recv() {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			msg := message
			tfGroup.Go(func() error {
				// Transformed messages are never returned to the caller because
				// invocation is asynchronous.
				if _, err := sub.Transform(tfCtx, msg); err != nil {
					return err
				}

				return nil
			})
		}

		if err := tfGroup.Wait(); err != nil {
			return err
		}

		// CTRL messages flush the pipeline. This must be done
		// after all messages have been processed.
		ctrl := message.New().AsControl()
		if _, err := sub.Transform(ctx, ctrl); err != nil {
			return err
		}

		return nil
	})

	// Data ingest.
	group.Go(func() error {
		defer ch.Close()

		eventSourceArn := event.Records[len(event.Records)-1].EventSourceArn
		converted := convertEventsRecords(event.Records)
		deaggregated, err := deaggregator.DeaggregateRecords(converted)
		if err != nil {
			return err
		}

		for _, record := range deaggregated {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			// Create Message metadata.
			m := kinesisStreamMetadata{
				*record.ApproximateArrivalTimestamp,
				eventSourceArn,
				*record.PartitionKey,
				*record.SequenceNumber,
			}

			metadata, err := json.Marshal(m)
			if err != nil {
				return err
			}

			msg := message.New().SetData(record.Data).SetMetadata(metadata).SkipMissingValues()
			ch.Send(msg)
		}

		return nil
	})

	// Wait for all goroutines to complete. This includes the goroutines that are
	// executing the transform functions.
	if err := group.Wait(); err != nil {
		return err
	}

	return nil
}

func convertEventsRecords(records []events.KinesisEventRecord) []types.Record {
	output := make([]types.Record, 0)

	for _, r := range records {
		// ApproximateArrivalTimestamp is events.SecondsEpochTime which serializes time.Time
		ts := r.Kinesis.ApproximateArrivalTimestamp.UTC()
		output = append(output, types.Record{
			ApproximateArrivalTimestamp: &ts,
			Data:                        r.Kinesis.Data,
			EncryptionType:              types.EncryptionType(r.Kinesis.EncryptionType),
			PartitionKey:                &r.Kinesis.PartitionKey,
			SequenceNumber:              &r.Kinesis.SequenceNumber,
		})
	}

	return output
}
