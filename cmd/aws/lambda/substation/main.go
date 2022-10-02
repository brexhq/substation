package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/kinesis-aggregation/go/deaggregator"
	"golang.org/x/sync/errgroup"

	"github.com/brexhq/substation/cmd"
	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/aws/appconfig"
	"github.com/brexhq/substation/internal/aws/kinesis"
	"github.com/brexhq/substation/internal/aws/s3manager"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/log"
)

var handler string
var scanMethod string

// errLambdaMissingHandler is returned when the Lambda is deployed without a configured handler.
const errLambdaMissingHandler = errors.Error("missing SUBSTATION_HANDLER environment variable")

// errLambdaInvalidHandler is returned when the Lambda is deployed with an unsupported handler.
const errLambdaInvalidHandler = errors.Error("invalid handler")

func main() {
	switch h := handler; h {
	case "AWS_API_GATEWAY":
		lambda.Start(gatewayHandler)
	case "AWS_KINESIS":
		lambda.Start(kinesisHandler)
	case "AWS_S3":
		lambda.Start(s3Handler)
	case "AWS_S3_SNS":
		lambda.Start(s3SnsHandler)
	case "AWS_SNS":
		lambda.Start(snsHandler)
	case "AWS_SQS":
		lambda.Start(sqsHandler)
	default:
		panic(fmt.Errorf("main handler %s: %v", h, errLambdaInvalidHandler))
	}
}

func init() {
	var found bool
	handler, found = os.LookupEnv("SUBSTATION_HANDLER")
	if !found {
		panic(fmt.Errorf("init handler %s: %v", handler, errLambdaMissingHandler))
	}

	// retrieves scan method from SUBSTATION_SCAN_METHOD environment variable
	scanMethod = cmd.GetScanMethod()
}

type gatewayMetadata struct {
	Resource string            `json:"resource"`
	Path     string            `json:"path"`
	Headers  map[string]string `json:"headers"`
}

func gatewayHandler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	sub := cmd.New()

	conf, err := appconfig.GetPrefetch(ctx)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500}, fmt.Errorf("gateway handler: %v", err)
	}
	json.Unmarshal(conf, &sub.Config)

	// maintains app state
	group, ctx := errgroup.WithContext(ctx)

	var sinkWg sync.WaitGroup
	sinkWg.Add(1)
	group.Go(func() error {
		return sub.Sink(ctx, &sinkWg)
	})

	var transformWg sync.WaitGroup
	for w := 0; w < sub.Concurrency(); w++ {
		transformWg.Add(1)
		group.Go(func() error {
			return sub.Transform(ctx, &transformWg)
		})
	}

	// ingest
	group.Go(func() error {
		if len(request.Body) != 0 {
			cap := config.NewCapsule()
			cap.SetData([]byte(request.Body))
			cap.SetMetadata(gatewayMetadata{
				request.Resource,
				request.Path,
				request.Headers,
			})

			sub.Send(cap)
		}

		sub.WaitTransform(&transformWg)
		sub.WaitSink(&sinkWg)

		return nil
	})

	if err := sub.Block(ctx, group); err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500}, fmt.Errorf("gateway handler: %v", err)
	}

	return events.APIGatewayProxyResponse{StatusCode: 200}, nil
}

type kinesisMetadata struct {
	ApproximateArrivalTimestamp time.Time `json:"approximateArrivalTimestamp"`
	EventSourceArn              string    `json:"eventSourceArn"`
	PartitionKey                string    `json:"partitionKey"`
	SequenceNumber              string    `json:"sequenceNumber"`
}

func kinesisHandler(ctx context.Context, event events.KinesisEvent) error {
	sub := cmd.New()

	conf, err := appconfig.GetPrefetch(ctx)
	if err != nil {
		return fmt.Errorf("kinesis handler: %v", err)
	}
	json.Unmarshal(conf, &sub.Config)

	// maintains app state
	group, ctx := errgroup.WithContext(ctx)

	var sinkWg sync.WaitGroup
	sinkWg.Add(1)
	group.Go(func() error {
		return sub.Sink(ctx, &sinkWg)
	})

	var transformWg sync.WaitGroup
	for w := 0; w < sub.Concurrency(); w++ {
		transformWg.Add(1)
		group.Go(func() error {
			return sub.Transform(ctx, &transformWg)
		})
	}

	// ingest
	eventSourceArn := event.Records[len(event.Records)-1].EventSourceArn

	group.Go(func() error {
		converted := kinesis.ConvertEventsRecords(event.Records)
		deaggregated, err := deaggregator.DeaggregateRecords(converted)
		if err != nil {
			return err
		}

		for _, record := range deaggregated {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				cap := config.NewCapsule()
				cap.SetData(record.Data)
				cap.SetMetadata(kinesisMetadata{
					*record.ApproximateArrivalTimestamp,
					eventSourceArn,
					*record.PartitionKey,
					*record.SequenceNumber,
				})

				sub.Send(cap)
			}
		}

		sub.WaitTransform(&transformWg)
		sub.WaitSink(&sinkWg)

		return nil
	})

	if err := sub.Block(ctx, group); err != nil {
		return fmt.Errorf("kinesis handler: %v", err)
	}

	return nil
}

type s3Metadata struct {
	EventTime  time.Time `json:"eventTime"`
	BucketArn  string    `json:"bucketArn"`
	BucketName string    `json:"bucketName"`
	ObjectKey  string    `json:"objectKey"`
	ObjectSize int64     `json:"objectSize"`
}

func s3Handler(ctx context.Context, event events.S3Event) error {
	sub := cmd.New()

	conf, err := appconfig.GetPrefetch(ctx)
	if err != nil {
		return fmt.Errorf("s3 handler: %v", err)
	}
	json.Unmarshal(conf, &sub.Config)

	// maintains app state
	group, ctx := errgroup.WithContext(ctx)

	var sinkWg sync.WaitGroup
	sinkWg.Add(1)
	group.Go(func() error {
		return sub.Sink(ctx, &sinkWg)
	})

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

			scanner, err := api.DownloadAsScanner(
				ctx,
				record.S3.Bucket.Name,
				record.S3.Object.Key,
			)
			if err != nil {
				return err
			}

			cap := config.NewCapsule()
			cap.SetMetadata(s3Metadata{
				record.EventTime,
				record.S3.Bucket.Arn,
				record.S3.Bucket.Name,
				record.S3.Object.Key,
				record.S3.Object.Size,
			})

			for scanner.Scan() {
				switch scanMethod {
				case "bytes":
					cap.SetData(scanner.Bytes())
				case "text":
					cap.SetData([]byte(scanner.Text()))
				}

				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
					sub.Send(cap)
				}
			}
		}

		sub.WaitTransform(&transformWg)
		sub.WaitSink(&sinkWg)

		return nil
	})

	if err := sub.Block(ctx, group); err != nil {
		return fmt.Errorf("s3 handler: %v", err)
	}

	return nil
}

func s3SnsHandler(ctx context.Context, event events.SNSEvent) error {
	sub := cmd.New()

	conf, err := appconfig.GetPrefetch(ctx)
	if err != nil {
		return fmt.Errorf("s3Sns handler: %v", err)
	}
	json.Unmarshal(conf, &sub.Config)

	// maintains app state
	group, ctx := errgroup.WithContext(ctx)

	var sinkWg sync.WaitGroup
	sinkWg.Add(1)
	group.Go(func() error {
		return sub.Sink(ctx, &sinkWg)
	})

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

				scanner, err := api.DownloadAsScanner(
					ctx,
					record.S3.Bucket.Name,
					record.S3.Object.Key,
				)
				if err != nil {
					return err
				}

				cap := config.NewCapsule()
				cap.SetMetadata(s3Metadata{
					record.EventTime,
					record.S3.Bucket.Arn,
					record.S3.Bucket.Name,
					record.S3.Object.Key,
					record.S3.Object.Size,
				})

				for scanner.Scan() {
					switch scanMethod {
					case "bytes":
						cap.SetData(scanner.Bytes())
					case "text":
						cap.SetData([]byte(scanner.Text()))
					}

					select {
					case <-ctx.Done():
						return ctx.Err()
					default:
						sub.Send(cap)
					}
				}
			}
		}

		sub.WaitTransform(&transformWg)
		sub.WaitSink(&sinkWg)

		return nil
	})

	if err := sub.Block(ctx, group); err != nil {
		return fmt.Errorf("s3 sns handler: %v", err)
	}

	return nil
}

type snsMetadata struct {
	Timestamp            time.Time `json:"timestamp"`
	EventSubscriptionArn string    `json:"eventSubscriptionArn"`
	MessageID            string    `json:"messageId"`
	Subject              string    `json:"subject"`
}

func snsHandler(ctx context.Context, event events.SNSEvent) error {
	sub := cmd.New()

	conf, err := appconfig.GetPrefetch(ctx)
	if err != nil {
		return fmt.Errorf("sns handler: %v", err)
	}
	json.Unmarshal(conf, &sub.Config)

	// maintains app state
	group, ctx := errgroup.WithContext(ctx)

	var sinkWg sync.WaitGroup
	sinkWg.Add(1)
	group.Go(func() error {
		return sub.Sink(ctx, &sinkWg)
	})

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
				cap := config.NewCapsule()
				cap.SetData([]byte(record.SNS.Message))
				cap.SetMetadata(snsMetadata{
					record.SNS.Timestamp,
					record.EventSubscriptionArn,
					record.SNS.MessageID,
					record.SNS.Subject,
				})

				sub.Send(cap)
			}
		}

		sub.WaitTransform(&transformWg)
		sub.WaitSink(&sinkWg)

		return nil
	})

	if err := sub.Block(ctx, group); err != nil {
		return fmt.Errorf("sns handler: %v", err)
	}

	return nil
}

type sqsMetadata struct {
	EventSourceArn string            `json:"eventSourceArn"`
	MessageID      string            `json:"messageId"`
	BodyMd5        string            `json:"bodyMd5"`
	Attributes     map[string]string `json:"attributes"`
}

func sqsHandler(ctx context.Context, event events.SQSEvent) error {
	sub := cmd.New()

	conf, err := appconfig.GetPrefetch(ctx)
	if err != nil {
		return fmt.Errorf("sqs handler: %v", err)
	}
	json.Unmarshal(conf, &sub.Config)

	// maintains app state
	group, ctx := errgroup.WithContext(ctx)

	var sinkWg sync.WaitGroup
	sinkWg.Add(1)
	group.Go(func() error {
		return sub.Sink(ctx, &sinkWg)
	})

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
				cap := config.NewCapsule()
				cap.SetData([]byte(msg.Body))
				cap.SetMetadata(sqsMetadata{
					msg.EventSourceARN,
					msg.MessageId,
					msg.Md5OfBody,
					msg.Attributes,
				})

				sub.Send(cap)
			}
		}

		sub.WaitTransform(&transformWg)
		sub.WaitSink(&sinkWg)

		return nil
	})

	if err := sub.Block(ctx, group); err != nil {
		return fmt.Errorf("sns handler: %v", err)
	}

	return nil
}
