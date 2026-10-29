package main

import (
	"context"
	"encoding/json"

	"golang.org/x/sync/errgroup"

	"github.com/brexhq/substation/v2"
	"github.com/brexhq/substation/v2/message"

	"github.com/brexhq/substation/v2/internal/channel"
)

type messageSender func(*message.Message)

func loadRuntime(ctx context.Context) (customConfig, *substation.Substation, error) {
	conf, err := getConfig(ctx)
	if err != nil {
		return customConfig{}, nil, err
	}

	cfg := customConfig{}
	if err := json.NewDecoder(conf).Decode(&cfg); err != nil {
		return customConfig{}, nil, err
	}

	// Catches an edge case where a missing concurrency value
	// can deadlock the application.
	if cfg.Concurrency == 0 {
		cfg.Concurrency = 1
	}

	sub, err := substation.New(ctx, cfg.Config)
	if err != nil {
		return customConfig{}, nil, err
	}

	return cfg, sub, nil
}

// runAsyncTransform executes transforms concurrently for messages produced by
// ingest. Transformed messages are not returned to the caller because
// invocation is asynchronous.
func runAsyncTransform(ctx context.Context, sub *substation.Substation, concurrency int, ingest func(ctx context.Context, send messageSender) error) error {
	ch := channel.New[*message.Message]()
	group, ctx := errgroup.WithContext(ctx)

	// Data transformation. Transforms are executed concurrently using a worker pool
	// managed by an errgroup. Each message is processed in a separate goroutine.
	group.Go(func() error {
		tfGroup, tfCtx := errgroup.WithContext(ctx)
		tfGroup.SetLimit(concurrency)

		for message := range ch.Recv() {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			msg := message
			tfGroup.Go(func() error {
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
		// after all messages have been processed, and with the outer context
		// because tfGroup.Wait cancels tfCtx.
		ctrl := message.New().AsControl()
		if _, err := sub.Transform(ctx, ctrl); err != nil {
			return err
		}

		return nil
	})

	// Data ingest.
	group.Go(func() error {
		defer ch.Close()

		return ingest(ctx, ch.Send)
	})

	return group.Wait()
}
