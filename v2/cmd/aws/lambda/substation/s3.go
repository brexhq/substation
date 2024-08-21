package main

import (
	"context"
	"encoding/json"
	"io"
	"net/url"
	"os"
	"slices"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/brexhq/substation/v2"
	"github.com/brexhq/substation/v2/internal/aws"
	"github.com/brexhq/substation/v2/internal/aws/s3manager"
	"github.com/brexhq/substation/v2/internal/bufio"
	"github.com/brexhq/substation/v2/internal/channel"
	"github.com/brexhq/substation/v2/internal/media"
	"github.com/brexhq/substation/v2/message"
	"golang.org/x/sync/errgroup"
)

type s3Metadata struct {
	EventTime  time.Time `json:"eventTime"`
	BucketArn  string    `json:"bucketArn"`
	BucketName string    `json:"bucketName"`
	ObjectKey  string    `json:"objectKey"`
	ObjectSize int64     `json:"objectSize"`
}

//nolint:gocognit
func s3Handler(ctx context.Context, event events.S3Event) error {
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

		client := s3manager.DownloaderAPI{}
		client.Setup(aws.Config{})

		for _, record := range event.Records {
			// The S3 object key is URL encoded.
			//
			// https://docs.aws.amazon.com/AmazonS3/latest/userguide/notification-content-structure.html
			objectKey, err := url.QueryUnescape(record.S3.Object.Key)
			if err != nil {
				return err
			}

			m := s3Metadata{
				EventTime:  record.EventTime,
				BucketArn:  record.S3.Bucket.Arn,
				BucketName: record.S3.Bucket.Name,
				ObjectKey:  objectKey,
				ObjectSize: record.S3.Object.Size,
			}

			metadata, err := json.Marshal(m)
			if err != nil {
				return err
			}

			dst, err := os.CreateTemp("", "substation")
			if err != nil {
				return err
			}
			defer os.Remove(dst.Name())
			defer dst.Close()

			if _, err := client.Download(ctx, record.S3.Bucket.Name, objectKey, dst); err != nil {
				return err
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

				msg := message.New().SetData(r).SetMetadata(metadata)
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
				msg := message.New().SetData(b).SetMetadata(metadata)

				ch.Send(msg)
			}

			if err := scanner.Err(); err != nil {
				return err
			}
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

//nolint:gocognit
func s3SnsHandler(ctx context.Context, event events.SNSEvent) error {
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

	// Data ingest.
	group.Go(func() error {
		defer ch.Close()

		client := s3manager.DownloaderAPI{}
		client.Setup(aws.Config{})

		for _, record := range event.Records {
			var s3Event events.S3Event
			err := json.Unmarshal([]byte(record.SNS.Message), &s3Event)
			if err != nil {
				return err
			}

			for _, record := range s3Event.Records {
				// The S3 object key is URL encoded.
				//
				// https://docs.aws.amazon.com/AmazonS3/latest/userguide/notification-content-structure.html
				objectKey, err := url.QueryUnescape(record.S3.Object.Key)
				if err != nil {
					return err
				}

				m := s3Metadata{
					record.EventTime,
					record.S3.Bucket.Arn,
					record.S3.Bucket.Name,
					objectKey,
					record.S3.Object.Size,
				}
				metadata, err := json.Marshal(m)
				if err != nil {
					return err
				}

				dst, err := os.CreateTemp("", "substation")
				if err != nil {
					return err
				}
				defer os.Remove(dst.Name())
				defer dst.Close()

				if _, err := client.Download(ctx, record.S3.Bucket.Name, objectKey, dst); err != nil {
					return err
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

					msg := message.New().SetData(r).SetMetadata(metadata)
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
					msg := message.New().SetData(b).SetMetadata(metadata)

					ch.Send(msg)
				}

				if err := scanner.Err(); err != nil {
					return err
				}
			}
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
