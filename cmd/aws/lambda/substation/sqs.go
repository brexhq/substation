package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-lambda-go/events"
	"github.com/brexhq/substation/cmd"
	"github.com/brexhq/substation/config"
	"golang.org/x/sync/errgroup"
)

type sqsMetadata struct {
	EventSourceArn string            `json:"eventSourceArn"`
	MessageID      string            `json:"messageId"`
	BodyMd5        string            `json:"bodyMd5"`
	Attributes     map[string]string `json:"attributes"`
}

func sqsHandler(ctx context.Context, event events.SQSEvent) error {
	sub := cmd.New()

	// retrieve and load configuration
	cfg, err := getConfig(ctx)
	if err != nil {
		return fmt.Errorf("sqs handler: %v", err)
	}

	if err := sub.SetConfig(cfg); err != nil {
		return fmt.Errorf("sqs handler: %v", err)
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
		for _, msg := range event.Records {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				capsule := config.NewCapsule()
				capsule.SetData([]byte(msg.Body))
				if _, err := capsule.SetMetadata(sqsMetadata{
					msg.EventSourceARN,
					msg.MessageId,
					msg.Md5OfBody,
					msg.Attributes,
				}); err != nil {
					return fmt.Errorf("sqs handler: %v", err)
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
		return fmt.Errorf("sqs handler: %v", err)
	}

	return nil
}
