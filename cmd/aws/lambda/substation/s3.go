package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"sync/atomic"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/brexhq/substation"
	"github.com/brexhq/substation/internal/aws"
	"github.com/brexhq/substation/internal/aws/s3manager"
	"github.com/brexhq/substation/internal/bufio"
	"github.com/brexhq/substation/internal/channel"
	"github.com/brexhq/substation/internal/metrics"
	mess "github.com/brexhq/substation/message"
	"github.com/brexhq/substation/transform"
	"golang.org/x/sync/errgroup"
)

type s3Metadata struct {
	EventTime  time.Time `json:"eventTime"`
	BucketArn  string    `json:"bucketArn"`
	BucketName string    `json:"bucketName"`
	ObjectKey  string    `json:"objectKey"`
	ObjectSize int64     `json:"objectSize"`
}

//nolint: gocognit // Ignore cognitive complexity.
func s3Handler(ctx context.Context, event events.S3Event) error {
	// Retrieve and load configuration.
	conf, err := getConfig(ctx)
	if err != nil {
		return fmt.Errorf("s3 handler: %v", err)
	}

	cfg := customConfig{}
	if err := json.NewDecoder(conf).Decode(&cfg); err != nil {
		return fmt.Errorf("s3 handler: %v", err)
	}

	sub, err := substation.New(ctx, cfg.Config)
	if err != nil {
		return fmt.Errorf("s3 handler: %v", err)
	}

	defer sub.Close(ctx)

	ch := channel.New[*mess.Message]()
	group, ctx := errgroup.WithContext(ctx)

	// Application metrics.
	var msgRecv, msgTran uint32
	metric, err := metrics.New(ctx, cfg.Metrics)
	if err != nil {
		return fmt.Errorf("s3 handler: %v", err)
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

		client := s3manager.DownloaderAPI{}

		// TODO: Add support for regions and credentials.
		client.Setup(aws.Config{})

		// Create Message metadata.
		m := s3Metadata{
			EventTime:  event.Records[0].EventTime,
			BucketArn:  event.Records[0].S3.Bucket.Arn,
			BucketName: event.Records[0].S3.Bucket.Name,
			ObjectKey:  event.Records[0].S3.Object.Key,
			ObjectSize: event.Records[0].S3.Object.Size,
		}

		metadata, err := json.Marshal(m)
		if err != nil {
			return fmt.Errorf("s3 handler: %v", err)
		}

		for _, record := range event.Records {
			// The S3 object key is URL encoded.
			//
			// https://docs.aws.amazon.com/AmazonS3/latest/userguide/notification-content-structure.html
			objectKey, err := url.QueryUnescape(record.S3.Object.Key)
			if err != nil {
				return fmt.Errorf("s3 handler: %v", err)
			}

			dst, err := os.CreateTemp("", "substation")
			if err != nil {
				return fmt.Errorf("s3 handler: %v", err)
			}
			defer os.Remove(dst.Name())
			defer dst.Close()

			if _, err := client.Download(ctx, record.S3.Bucket.Name, objectKey, dst); err != nil {
				return fmt.Errorf("s3 handler: %v", err)
			}

			scanner := bufio.NewScanner()
			defer scanner.Close()

			if err := scanner.ReadFile(dst); err != nil {
				return fmt.Errorf("s3 handler: %v", err)
			}

			for scanner.Scan() {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
				}

				msg, err := mess.New(
					mess.SetData([]byte(scanner.Text())),
					mess.SetMetadata(metadata),
				)
				if err != nil {
					return fmt.Errorf("s3 handler: %v", err)
				}

				ch.Send(msg)
				atomic.AddUint32(&msgRecv, 1)
			}
		}

		return nil
	})

	// Wait for all goroutines to complete. This includes the goroutines that are
	// executing the transform functions.
	if err := group.Wait(); err != nil {
		return fmt.Errorf("s3 handler: %v", err)
	}

	// Generate metrics.
	if err := metric.Generate(ctx, metrics.Data{
		Name:  "MessagesReceived",
		Value: msgRecv,
		Attributes: map[string]string{
			"FunctionName": functionName,
		},
	}); err != nil {
		return fmt.Errorf("s3 handler: %v", err)
	}

	if err := metric.Generate(ctx, metrics.Data{
		Name:  "MessagesTransformed",
		Value: msgTran,
		Attributes: map[string]string{
			"FunctionName": functionName,
		},
	}); err != nil {
		return fmt.Errorf("s3 handler: %v", err)
	}

	return nil
}

//nolint:gocognit
func s3SnsHandler(ctx context.Context, event events.SNSEvent) error {
	// Retrieve and load configuration.
	conf, err := getConfig(ctx)
	if err != nil {
		return fmt.Errorf("s3 handler: %v", err)
	}

	cfg := customConfig{}
	if err := json.NewDecoder(conf).Decode(&cfg); err != nil {
		return fmt.Errorf("s3 handler: %v", err)
	}

	sub, err := substation.New(ctx, cfg.Config)
	if err != nil {
		return fmt.Errorf("s3 handler: %v", err)
	}

	defer sub.Close(ctx)

	ch := channel.New[*mess.Message]()
	group, ctx := errgroup.WithContext(ctx)

	// Application metrics.
	var msgRecv, msgTran uint32
	metric, err := metrics.New(ctx, cfg.Metrics)
	if err != nil {
		return fmt.Errorf("s3 handler: %v", err)
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
					return fmt.Errorf("s3 handler: %v", err)
				}

				dst, err := os.CreateTemp("", "substation")
				if err != nil {
					return fmt.Errorf("s3 handler: %v", err)
				}
				defer os.Remove(dst.Name())
				defer dst.Close()

				if _, err := client.Download(ctx, record.S3.Bucket.Name, objectKey, dst); err != nil {
					return fmt.Errorf("s3 handler: %v", err)
				}

				scanner := bufio.NewScanner()
				defer scanner.Close()

				if err := scanner.ReadFile(dst); err != nil {
					return fmt.Errorf("s3 handler: %v", err)
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
					return fmt.Errorf("s3 handler: %v", err)
				}

				for scanner.Scan() {
					select {
					case <-ctx.Done():
						return ctx.Err()
					default:
					}

					msg, err := mess.New(
						mess.SetData([]byte(scanner.Text())),
						mess.SetMetadata(metadata),
					)
					if err != nil {
						return fmt.Errorf("s3 handler: %v", err)
					}

					ch.Send(msg)
					atomic.AddUint32(&msgRecv, 1)
				}
			}
		}

		return nil
	})

	// Wait for all goroutines to complete. This includes the goroutines that are
	// executing the transform functions.
	if err := group.Wait(); err != nil {
		return fmt.Errorf("s3 handler: %v", err)
	}

	// Generate metrics.
	if err := metric.Generate(ctx, metrics.Data{
		Name:  "MessagesReceived",
		Value: msgRecv,
		Attributes: map[string]string{
			"FunctionName": functionName,
		},
	}); err != nil {
		return fmt.Errorf("s3 handler: %v", err)
	}

	if err := metric.Generate(ctx, metrics.Data{
		Name:  "MessagesTransformed",
		Value: msgTran,
		Attributes: map[string]string{
			"FunctionName": functionName,
		},
	}); err != nil {
		return fmt.Errorf("s3 handler: %v", err)
	}

	return nil
}
