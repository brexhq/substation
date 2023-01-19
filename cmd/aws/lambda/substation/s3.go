package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/brexhq/substation/cmd"
	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/aws/s3manager"
	"github.com/brexhq/substation/internal/bufio"
	"github.com/brexhq/substation/internal/log"
	"golang.org/x/sync/errgroup"
)

type s3Metadata struct {
	EventTime  time.Time `json:"eventTime"`
	BucketArn  string    `json:"bucketArn"`
	BucketName string    `json:"bucketName"`
	ObjectKey  string    `json:"objectKey"`
	ObjectSize int64     `json:"objectSize"`
}

func s3Handler(ctx context.Context, event events.S3Event) error {
	sub := cmd.New()

	// retrieve and load configuration
	cfg, err := getConfig(ctx)
	if err != nil {
		return fmt.Errorf("s3 handler: %v", err)
	}

	if err := sub.SetConfig(cfg); err != nil {
		return fmt.Errorf("s3 handler: %v", err)
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
	api := s3manager.DownloaderAPI{}
	api.Setup()

	group.Go(func() error {
		for _, record := range event.Records {
			log.WithField(
				"bucket", record.S3.Bucket.Name,
			).WithField(
				"key", record.S3.Object.Key,
			).Debug("received S3 trigger")

			dst, err := os.CreateTemp("", "substation")
			if err != nil {
				return fmt.Errorf("s3 handler: %v", err)
			}
			defer os.Remove(dst.Name())
			defer dst.Close()

			if _, err := api.Download(ctx, record.S3.Bucket.Name, record.S3.Object.Key, dst); err != nil {
				return fmt.Errorf("s3 handler: %v", err)
			}

			capsule := config.NewCapsule()
			if _, err = capsule.SetMetadata(s3Metadata{
				record.EventTime,
				record.S3.Bucket.Arn,
				record.S3.Bucket.Name,
				record.S3.Object.Key,
				record.S3.Object.Size,
			}); err != nil {
				return fmt.Errorf("s3 handler: %v", err)
			}

			scanner := bufio.NewScanner()
			defer scanner.Close()

			if err := scanner.ReadFile(dst); err != nil {
				return fmt.Errorf("s3 handler: %v", err)
			}

			for scanner.Scan() {
				switch scanner.Method() {
				case "bytes":
					capsule.SetData(scanner.Bytes())
				case "text":
					capsule.SetData([]byte(scanner.Text()))
				}

				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
					sub.Send(capsule)
				}
			}
		}

		sub.WaitTransform(&transformWg)
		sub.WaitSink(&sinkWg)

		return nil
	})

	// block until ITL is complete
	if err := sub.Block(ctx, group); err != nil {
		return fmt.Errorf("s3 handler: %v", err)
	}

	return nil
}

func s3SnsHandler(ctx context.Context, event events.SNSEvent) error {
	sub := cmd.New()

	// retrieve and load configuration
	cfg, err := getConfig(ctx)
	if err != nil {
		return fmt.Errorf("s3 sns handler: %v", err)
	}

	if err := sub.SetConfig(cfg); err != nil {
		return fmt.Errorf("s3 sns handler: %v", err)
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
	api := s3manager.DownloaderAPI{}
	api.Setup()

	group.Go(func() error {
		for _, record := range event.Records {
			var s3Event events.S3Event
			err := json.Unmarshal([]byte(record.SNS.Message), &s3Event)
			if err != nil {
				return err
			}

			for _, record := range s3Event.Records {
				log.WithField(
					"bucket", record.S3.Bucket.Name,
				).WithField(
					"key", record.S3.Object.Key,
				).Debug("received S3 trigger")

				dst, err := os.CreateTemp("", "substation")
				if err != nil {
					return fmt.Errorf("s3 sns handler: %v", err)
				}
				defer os.Remove(dst.Name())
				defer dst.Close()

				if _, err := api.Download(ctx, record.S3.Bucket.Name, record.S3.Object.Key, dst); err != nil {
					return fmt.Errorf("s3 sns handler: %v", err)
				}

				capsule := config.NewCapsule()
				if _, err := capsule.SetMetadata(s3Metadata{
					record.EventTime,
					record.S3.Bucket.Arn,
					record.S3.Bucket.Name,
					record.S3.Object.Key,
					record.S3.Object.Size,
				}); err != nil {
					return fmt.Errorf("s3 sns handler: %v", err)
				}

				scanner := bufio.NewScanner()
				defer scanner.Close()

				if err := scanner.ReadFile(dst); err != nil {
					return fmt.Errorf("s3 sns handler: %v", err)
				}

				for scanner.Scan() {
					switch scanner.Method() {
					case "bytes":
						capsule.SetData(scanner.Bytes())
					case "text":
						capsule.SetData([]byte(scanner.Text()))
					}

					select {
					case <-ctx.Done():
						return ctx.Err()
					default:
						sub.Send(capsule)
					}
				}
			}
		}

		sub.WaitTransform(&transformWg)
		sub.WaitSink(&sinkWg)

		return nil
	})

	// block until ITL is complete
	if err := sub.Block(ctx, group); err != nil {
		return fmt.Errorf("s3 sns handler: %v", err)
	}

	return nil
}
