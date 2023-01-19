package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/awslabs/kinesis-aggregation/go/deaggregator"
	"github.com/brexhq/substation/cmd"
	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/aws/kinesis"
	"golang.org/x/sync/errgroup"
)

type kinesisMetadata struct {
	ApproximateArrivalTimestamp time.Time `json:"approximateArrivalTimestamp"`
	EventSourceArn              string    `json:"eventSourceArn"`
	PartitionKey                string    `json:"partitionKey"`
	SequenceNumber              string    `json:"sequenceNumber"`
}

func kinesisHandler(ctx context.Context, event events.KinesisEvent) error {
	sub := cmd.New()

	// retrieve and load configuration
	cfg, err := getConfig(ctx)
	if err != nil {
		return fmt.Errorf("kinesis handler: %v", err)
	}

	if err := sub.SetConfig(cfg); err != nil {
		return fmt.Errorf("kinesis handler: %v", err)
	}

	// maintains app state
	group, ctx := errgroup.WithContext(ctx)

	// load
	var sinkWg sync.WaitGroup
	sinkWg.Add(1)
	group.Go(func() error {
		return sub.Sink(ctx, &sinkWg)
	})

	// transform
	var transformWg sync.WaitGroup
	for w := 0; w < sub.Concurrency(); w++ {
		transformWg.Add(1)
		group.Go(func() error {
			return sub.Transform(ctx, &transformWg)
		})
	}

	// ingest
	eventSourceArn := event.Records[len(event.Records)-1].EventSourceArn

	group.Go(func() error {
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
				capsule := config.NewCapsule()
				capsule.SetData(record.Data)
				if _, err := capsule.SetMetadata(kinesisMetadata{
					*record.ApproximateArrivalTimestamp,
					eventSourceArn,
					*record.PartitionKey,
					*record.SequenceNumber,
				}); err != nil {
					return fmt.Errorf("kinesis handler: %v", err)
				}

				sub.Send(capsule)
			}
		}

		sub.WaitTransform(&transformWg)
		sub.WaitSink(&sinkWg)

		return nil
	})

	// block until ITL is complete
	if err := sub.Block(ctx, group); err != nil {
		return fmt.Errorf("kinesis handler: %v", err)
	}

	return nil
}
