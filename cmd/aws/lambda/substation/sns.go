package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/brexhq/substation/cmd"
	"github.com/brexhq/substation/config"
	"golang.org/x/sync/errgroup"
)

type snsMetadata struct {
	Timestamp            time.Time `json:"timestamp"`
	EventSubscriptionArn string    `json:"eventSubscriptionArn"`
	MessageID            string    `json:"messageId"`
	Subject              string    `json:"subject"`
}

func snsHandler(ctx context.Context, event events.SNSEvent) error {
	sub := cmd.New()

	// retrieve and load configuration
	cfg, err := getConfig(ctx)
	if err != nil {
		return fmt.Errorf("sns handler: %v", err)
	}

	if err := sub.SetConfig(cfg); err != nil {
		return fmt.Errorf("sns handler: %v", err)
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
	group.Go(func() error {
		for _, record := range event.Records {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				capsule := config.NewCapsule()
				capsule.SetData([]byte(record.SNS.Message))
				if _, err := capsule.SetMetadata(snsMetadata{
					record.SNS.Timestamp,
					record.EventSubscriptionArn,
					record.SNS.MessageID,
					record.SNS.Subject,
				}); err != nil {
					return fmt.Errorf("sns handler: %v", err)
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
		return fmt.Errorf("sns handler: %v", err)
	}

	return nil
}
