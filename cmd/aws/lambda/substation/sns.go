package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/brexhq/substation"
	"github.com/brexhq/substation/internal/channel"
	"github.com/brexhq/substation/message"
	"github.com/brexhq/substation/transform"
	"golang.org/x/sync/errgroup"
)

type snsMetadata struct {
	Timestamp            time.Time `json:"timestamp"`
	EventSubscriptionArn string    `json:"eventSubscriptionArn"`
	MessageID            string    `json:"messageId"`
	Subject              string    `json:"subject"`
}

func snsHandler(ctx context.Context, event events.SNSEvent) error {
	// Retrieve and load configuration.
	conf, err := getConfig(ctx)
	if err != nil {
		return fmt.Errorf("sns handler: %v", err)
	}

	cfg := customConfig{}
	if err := json.NewDecoder(conf).Decode(&cfg); err != nil {
		return fmt.Errorf("sns handler: %v", err)
	}

	sub, err := substation.New(ctx, cfg.Config)
	if err != nil {
		return fmt.Errorf("sns handler: %v", err)
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
				// Transformed messages are never returned to the caller because
				// invocation is asynchronous.
				if _, err := transform.Apply(tfCtx, sub.Transforms(), m); err != nil {
					return err
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

		// Create Message metadata.
		m := snsMetadata{
			Timestamp:            event.Records[0].SNS.Timestamp,
			EventSubscriptionArn: event.Records[0].EventSubscriptionArn,
			MessageID:            event.Records[0].SNS.MessageID,
			Subject:              event.Records[0].SNS.Subject,
		}

		metadata, err := json.Marshal(m)
		if err != nil {
			return fmt.Errorf("sns handler: %v", err)
		}

		for _, record := range event.Records {
			b := []byte(record.SNS.Message)
			msg := message.New().SetData(b).SetMetadata(metadata)

			ch.Send(msg)
		}

		return nil
	})

	// Wait for all goroutines to complete. This includes the goroutines that are
	// executing the transform functions.
	if err := group.Wait(); err != nil {
		return fmt.Errorf("sns handler: %v", err)
	}

	return nil
}
