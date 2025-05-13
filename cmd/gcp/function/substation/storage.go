package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"slices"

	"cloud.google.com/go/storage"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"golang.org/x/sync/errgroup"

	"github.com/brexhq/substation/v2"
	"github.com/brexhq/substation/v2/message"

	"github.com/brexhq/substation/v2/internal/bufio"
	"github.com/brexhq/substation/v2/internal/channel"
	"github.com/brexhq/substation/v2/internal/media"
)

type CloudStorageEvent struct {
	Bucket string `json:"bucket"`
	Name   string `json:"name"`
	Size   int64  `json:"size"`
	Time   string `json:"time"`
}

func cloudStorageHandler(ctx context.Context, e cloudevents.Event) error {
	// Retrieve and load configuration.
	conf, err := getConfig(ctx)
	if err != nil {
		return err
	}

	cfg := customConfig{}
	if err := json.NewDecoder(conf).Decode(&cfg); err != nil {
		return err
	}

	// Catches an edge case where a missing concurrency value
	// can deadlock the application.
	if cfg.Concurrency == 0 {
		cfg.Concurrency = 1
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
		if _, err := sub.Transform(tfCtx, ctrl); err != nil {
			return err
		}

		return nil
	})

	// Data ingest
	group.Go(func() error {
		defer ch.Close()

		var evt CloudStorageEvent
		if err := json.Unmarshal(e.Data(), &evt); err != nil {
			return fmt.Errorf("failed to unmarshal event data: %v", err)
		}

		// Create a storage client
		client, err := storage.NewClient(ctx)
		if err != nil {
			return fmt.Errorf("storage.NewClient: %v", err)
		}
		defer client.Close()

		reader, err := client.Bucket(evt.Bucket).Object(evt.Name).NewReader(ctx)
		if err != nil {
			return fmt.Errorf("Object(%q).NewReader: %v", evt.Name, err)
		}
		defer reader.Close()

		dst, err := os.CreateTemp("", "substation")
		if err != nil {
			return err
		}
		defer os.Remove(dst.Name())
		defer dst.Close()

		if _, err := io.Copy(dst, reader); err != nil {
			return fmt.Errorf("io.Copy: %w", err)
		}

		// Determines if the file should be treated as text.
		// Text files are decompressed by the bufio package
		// (if necessary) and each line is sent as a separate
		// message. All other files are sent as a single message.
		mediaType, err := media.File(dst)
		if err != nil {
			return err
		}

		if _, err := dst.Seek(0, 0); err != nil {
			return err
		}

		// Unsupported media types are sent as binary data.
		if !slices.Contains(bufio.MediaTypes, mediaType) {
			r, err := io.ReadAll(dst)
			if err != nil {
				return err
			}

			msg := message.New().SetData(r)
			ch.Send(msg)

			return nil
		}

		scanner := bufio.NewScanner()
		defer scanner.Close()

		if err := scanner.ReadFile(dst); err != nil {
			return err
		}

		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			b := []byte(scanner.Text())
			msg := message.New().SetData(b)

			ch.Send(msg)
		}

		if err := scanner.Err(); err != nil {
			return err
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
