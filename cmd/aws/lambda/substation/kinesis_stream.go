package main

import (
	"context"
	"encoding/json"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/awslabs/kinesis-aggregation/go/deaggregator"
	"github.com/brexhq/substation"
	"github.com/brexhq/substation/internal/aws/kinesis"
	"github.com/brexhq/substation/internal/channel"
	"github.com/brexhq/substation/message"
	"github.com/brexhq/substation/transform"
	"golang.org/x/sync/errgroup"
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
	// managed by an errgroup. Each Message is processed in a separate goroutine.
	group.Go(func() error {
		tfGroup, tfCtx := errgroup.WithContext(ctx)
		tfGroup.SetLimit(cfg.Concurrency)

		for message := range ch.Recv() {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			m := message
			tfGroup.Go(func() error {
				msg, err := transform.Apply(tfCtx, sub.Transforms(), m)
				if err != nil {
					return err
				}

				for _, m := range msg {
					if m.IsControl() {
						continue
					}
				}

				return nil
			})
		}

		if err := tfGroup.Wait(); err != nil {
			return err
		}

		// CTRL Messages flush the transform functions. This must be done
		// after all messages have been processed.
		ctrl := message.New(message.AsControl())
		if _, err := transform.Apply(ctx, sub.Transforms(), ctrl); err != nil {
			return err
		}

		return nil
	})

	// Data ingest. A CTRL Message is sent to the transforms after all data has been
	// sent to the channel.
	group.Go(func() error {
		defer ch.Close()

		eventSourceArn := event.Records[len(event.Records)-1].EventSourceArn
		converted := kinesis.ConvertEventsRecords(event.Records)
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

			msg := message.New().SetData(record.Data).SetMetadata(metadata)
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
