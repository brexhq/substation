package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/brexhq/substation"
	"github.com/brexhq/substation/internal/channel"
	"github.com/brexhq/substation/internal/metrics"
	mess "github.com/brexhq/substation/message"
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

	defer sub.Close(ctx)

	ch := channel.New[*mess.Message]()
	group, ctx := errgroup.WithContext(ctx)

	// Application metrics.
	var msgRecv, msgTran uint32
	metric, err := metrics.New(ctx, cfg.Metrics)
	if err != nil {
		return fmt.Errorf("sns handler: %v", err)
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
			message, err := mess.New(
				mess.SetData([]byte(record.SNS.Message)),
				mess.SetMetadata(metadata),
			)
			if err != nil {
				return err
			}

			ch.Send(message)
			atomic.AddUint32(&msgRecv, 1)
		}

		return nil
	})

	// Wait for all goroutines to complete. This includes the goroutines that are
	// executing the transform functions.
	if err := group.Wait(); err != nil {
		return fmt.Errorf("sns handler: %v", err)
	}

	// Generate metrics.
	if err := metric.Generate(ctx, metrics.Data{
		Name:  "MessagesReceived",
		Value: msgRecv,
		Attributes: map[string]string{
			"FunctionName": functionName,
		},
	}); err != nil {
		return fmt.Errorf("sns handler: %v", err)
	}

	if err := metric.Generate(ctx, metrics.Data{
		Name:  "MessagesTransformed",
		Value: msgTran,
		Attributes: map[string]string{
			"FunctionName": functionName,
		},
	}); err != nil {
		return fmt.Errorf("sns handler: %v", err)
	}

	return nil
}
