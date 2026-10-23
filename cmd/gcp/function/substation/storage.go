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

	"github.com/brexhq/substation/v2/message"

	"github.com/brexhq/substation/v2/internal/bufio"
	"github.com/brexhq/substation/v2/internal/media"
)

// CloudStorageEvent is the event data for a Cloud Storage event.
// This also provides metadata about the object that was created.
type CloudStorageEvent struct {
	Bucket string `json:"bucket"`
	Name   string `json:"name"`
	Size   string `json:"size"`
	Time   string `json:"time"`
}

func cloudStorageHandler(ctx context.Context, e cloudevents.Event) error {
	cfg, sub, err := loadRuntime(ctx)
	if err != nil {
		return err
	}

	return runAsyncTransform(ctx, sub, cfg.Concurrency, func(ctx context.Context, send messageSender) error {
		var evt CloudStorageEvent
		if err := json.Unmarshal(e.Data(), &evt); err != nil {
			return fmt.Errorf("failed to unmarshal event data: %v", err)
		}

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

		metadata, err := json.Marshal(evt)
		if err != nil {
			return err
		}

		// Unsupported media types are sent as binary data.
		if !slices.Contains(bufio.MediaTypes, mediaType) {
			r, err := io.ReadAll(dst)
			if err != nil {
				return err
			}

			send(message.New().SetData(r).SetMetadata(metadata))

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
			send(message.New().SetData(b).SetMetadata(metadata))
		}

		return scanner.Err()
	})
}
