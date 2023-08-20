package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/awslabs/kinesis-aggregation/go/deaggregator"
	"github.com/brexhq/substation"
	"github.com/brexhq/substation/internal/aws/kinesis"
	"github.com/brexhq/substation/internal/channel"
	"github.com/brexhq/substation/internal/metrics"
	mess "github.com/brexhq/substation/message"
	"github.com/brexhq/substation/transform"
	"golang.org/x/sync/errgroup"
)

type kinesisMetadata struct {
	ApproximateArrivalTimestamp time.Time `json:"approximateArrivalTimestamp"`
	Stream                      string    `json:"stream"`
	PartitionKey                string    `json:"partitionKey"`
	SequenceNumber              string    `json:"sequenceNumber"`
}

func kinesisHandler(ctx context.Context, event events.KinesisEvent) error {
	// Retrieve and load configuration.
	conf, err := getConfig(ctx)
	if err != nil {
		return fmt.Errorf("kinesis handler: %v", err)
	}

	cfg := customConfig{}
	if err := json.NewDecoder(conf).Decode(&cfg); err != nil {
		return fmt.Errorf("kinesis handler: %v", err)
	}

	sub, err := substation.New(ctx, cfg.Config)
	if err != nil {
		return fmt.Errorf("kinesis handler: %v", err)
	}
	defer sub.Close(ctx)

	ch := channel.New[*mess.Message]()
	group, ctx := errgroup.WithContext(ctx)

	// Application metrics.
	var msgRecv, msgTran uint32
	metric, err := metrics.New(ctx, cfg.Metrics)
	if err != nil {
		return fmt.Errorf("kinesis handler: %v", err)
	}

	// Data transformation. Transforms are executed concurrently using a worker pool
	// managed by an errgroup. Each message is processed in a separate goroutine.
	group.Go(func() error {
		group, ctx := errgroup.WithContext(ctx)
		group.SetLimit(cfg.Concurrency)

		for message := range ch.Recv() {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			m := message
			group.Go(func() error {
				msg, err := transform.Apply(ctx, sub.Transforms(), m)
				if err != nil {
					return err
				}

				for _, m := range msg {
					if m.IsControl() {
						continue
					}

					atomic.AddUint32(&msgTran, 1)
				}

				return nil
			})
		}

		if err := group.Wait(); err != nil {
			return err
		}

		// CTRL message is used to flush the transform functions. This must be done
		// after all messages have been processed.
		ctrl, err := mess.New(mess.AsControl())
		if err != nil {
			return err
		}

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
			return fmt.Errorf("kinesis handler: %v", err)
		}

		for _, record := range deaggregated {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			// Create Message metadata.
			m := kinesisMetadata{
				*record.ApproximateArrivalTimestamp,
				eventSourceArn,
				*record.PartitionKey,
				*record.SequenceNumber,
			}

			metadata, err := json.Marshal(m)
			if err != nil {
				return fmt.Errorf("kinesis handler: %v", err)
			}

			message, err := mess.New(
				mess.SetData(record.Data),
				mess.SetMetadata(metadata),
			)
			if err != nil {
				return fmt.Errorf("kinesis handler: %v", err)
			}

			ch.Send(message)
			atomic.AddUint32(&msgRecv, 1)
		}

		return nil
	})

	// Wait for all goroutines to complete. This includes the goroutines that are
	// executing the transform functions.
	if err := group.Wait(); err != nil {
		return fmt.Errorf("kinesis handler: %v", err)
	}

	// Generate metrics.
	if err := metric.Generate(ctx, metrics.Data{
		Name:  "MessagesReceived",
		Value: msgRecv,
		Attributes: map[string]string{
			"FunctionName": functionName,
		},
	}); err != nil {
		return fmt.Errorf("kinesis handler: %v", err)
	}

	if err := metric.Generate(ctx, metrics.Data{
		Name:  "MessagesTransformed",
		Value: msgTran,
		Attributes: map[string]string{
			"FunctionName": functionName,
		},
	}); err != nil {
		return fmt.Errorf("kinesis handler: %v", err)
	}

	return nil
}
