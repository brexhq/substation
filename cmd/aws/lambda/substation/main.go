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

	"github.com/brexhq/substation/cmd"
	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/aws/appconfig"
	"github.com/brexhq/substation/internal/aws/kinesis"
	"github.com/brexhq/substation/internal/aws/s3manager"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/log"
)

var sub cmd.Substation
var concurrency int
var handler string

// LambdaMissingHandler is returned when the Lambda is deployed without a configured handler.
const LambdaMissingHandler = errors.Error("LambdaMissingHandler")

// LambdaUnsupportedHandler is returned when the Lambda is deployed without a supported handler.
const LambdaUnsupportedHandler = errors.Error("LambdaUnsupportedHandler")

func main() {
	switch h := handler; h {
	case "GATEWAY":
		lambda.Start(gatewayHandler)
	case "KINESIS":
		lambda.Start(kinesisHandler)
	case "S3":
		lambda.Start(s3Handler)
	case "S3-SNS":
		lambda.Start(s3SnsHandler)
	case "SNS":
		lambda.Start(snsHandler)
	case "SQS":
		lambda.Start(sqsHandler)
	default:
		panic(fmt.Errorf("main handler %s: %v", h, LambdaUnsupportedHandler))
	}
}

func init() {
	var found bool
	handler, found = os.LookupEnv("SUBSTATION_HANDLER")
	if !found {
		panic(fmt.Errorf("init handler %s: %v", handler, LambdaMissingHandler))
	}

	// retrieves concurrency value from SUBSTATION_CONCURRENCY environment variable
	var err error
	concurrency, err = cmd.GetConcurrency()
	if err != nil {
		panic(fmt.Errorf("init concurrency: %v", err))
	}
}

type gatewayMetadata struct {
	Resource string            `json:"resource"`
	Path     string            `json:"path"`
	Headers  map[string]string `json:"headers"`
}

func gatewayHandler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	conf, err := appconfig.GetPrefetch(ctx)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500}, fmt.Errorf("gateway handler: %v", err)
	}
	json.Unmarshal(conf, &sub.Config)

	sub.CreateChannels(concurrency)
	defer sub.KillSignal()

	go func() {
		var sinkWg sync.WaitGroup
		var transformWg sync.WaitGroup

		sinkWg.Add(1)
		go sub.Sink(ctx, &sinkWg)

		transformWg.Add(1)
		go sub.Transform(ctx, &transformWg)

		if len(request.Body) != 0 {
			cap := config.NewCapsule()
			cap.SetData([]byte(request.Body))
			cap.SetMetadata(gatewayMetadata{
				request.Resource,
				request.Path,
				request.Headers,
			})

			sub.SendTransform(cap)
		}

		sub.TransformSignal()
		transformWg.Wait()
		sub.SinkSignal()
		sinkWg.Wait()
	}()

	if err := sub.Block(ctx); err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500}, fmt.Errorf("gateway handler: %v", err)
	}

	return events.APIGatewayProxyResponse{StatusCode: 200}, nil
}

type kinesisMetadata struct {
	EventSourceArn              string    `json:"event_source_arn"`
	ApproximateArrivalTimestamp time.Time `json:"approximate_arrival_timestamp"`
	PartitionKey                string    `json:"partition_key"`
	SequenceNumber              string    `json:"sequence_number"`
}

func kinesisHandler(ctx context.Context, event events.KinesisEvent) error {
	conf, err := appconfig.GetPrefetch(ctx)
	if err != nil {
		return fmt.Errorf("kinesis handler: %v", err)
	}
	json.Unmarshal(conf, &sub.Config)

	sub.CreateChannels(concurrency)
	defer sub.KillSignal()

	eventSourceArn := event.Records[len(event.Records)-1].EventSourceArn

	go func() {
		var sinkWg sync.WaitGroup
		var transformWg sync.WaitGroup

		sinkWg.Add(1)
		go sub.Sink(ctx, &sinkWg)

		for w := 0; w <= concurrency; w++ {
			transformWg.Add(1)
			go sub.Transform(ctx, &transformWg)
		}

		converted := kinesis.ConvertEventsRecords(event.Records)
		deaggregated, err := deaggregator.DeaggregateRecords(converted)
		if err != nil {
			sub.SendErr(fmt.Errorf("kinesis handler: %v", err))
			return
		}

		for _, record := range deaggregated {
			cap := config.NewCapsule()
			cap.SetData(record.Data)
			cap.SetMetadata(kinesisMetadata{
				eventSourceArn,
				*record.ApproximateArrivalTimestamp,
				*record.PartitionKey,
				*record.SequenceNumber,
			})

			sub.SendTransform(cap)
		}

		sub.TransformSignal()
		transformWg.Wait()
		sub.SinkSignal()
		sinkWg.Wait()
	}()

	if err := sub.Block(ctx); err != nil {
		return fmt.Errorf("kinesis handler: %v", err)
	}

	return nil
}

type s3Metadata struct {
	BucketArn  string `json:"bucket_arn"`
	BucketName string `json:"bucket_name"`
	ObjectKey  string `json:"object_key"`
	ObjectSize int64  `json:"object_size"`
}

func s3Handler(ctx context.Context, event events.S3Event) error {
	conf, err := appconfig.GetPrefetch(ctx)
	if err != nil {
		return fmt.Errorf("s3 handler: %v", err)
	}
	json.Unmarshal(conf, &sub.Config)

	sub.CreateChannels(concurrency)
	defer sub.KillSignal()

	go func() {
		api := s3manager.DownloaderAPI{}
		api.Setup()

		var sinkWg sync.WaitGroup
		var transformWg sync.WaitGroup

		sinkWg.Add(1)
		go sub.Sink(ctx, &sinkWg)

		for w := 0; w <= concurrency; w++ {
			transformWg.Add(1)
			go sub.Transform(ctx, &transformWg)
		}

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
				sub.SendErr(fmt.Errorf("s3 handler: %v", err))
				return
			}

			cap := config.NewCapsule()
			cap.SetMetadata(s3Metadata{
				record.S3.Bucket.Arn,
				record.S3.Bucket.Name,
				record.S3.Object.Key,
				record.S3.Object.Size,
			})

			for scanner.Scan() {
				cap.SetData([]byte(scanner.Text()))
				sub.SendTransform(cap)
			}
		}

		sub.TransformSignal()
		transformWg.Wait()
		sub.SinkSignal()
		sinkWg.Wait()
	}()

	if err := sub.Block(ctx); err != nil {
		return fmt.Errorf("s3 handler: %v", err)
	}

	return nil
}

func s3SnsHandler(ctx context.Context, event events.SNSEvent) error {
	conf, err := appconfig.GetPrefetch(ctx)
	if err != nil {
		return fmt.Errorf("sns-s3 handler: %v", err)
	}
	json.Unmarshal(conf, &sub.Config)

	sub.CreateChannels(concurrency)
	defer sub.KillSignal()

	go func() {
		api := s3manager.DownloaderAPI{}
		api.Setup()

		var sinkWg sync.WaitGroup
		var transformWg sync.WaitGroup

		sinkWg.Add(1)
		go sub.Sink(ctx, &sinkWg)

		for w := 0; w <= concurrency; w++ {
			transformWg.Add(1)
			go sub.Transform(ctx, &transformWg)
		}

		for _, record := range event.Records {
			var s3Event events.S3Event
			err := json.Unmarshal([]byte(record.SNS.Message), &s3Event)
			if err != nil {
				sub.SendErr(fmt.Errorf("sns-s3 handler: %v", err))
				return
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
					sub.SendErr(fmt.Errorf("sns-s3 handler: %v", err))
					return
				}

				cap := config.NewCapsule()
				cap.SetMetadata(s3Metadata{
					record.S3.Bucket.Arn,
					record.S3.Bucket.Name,
					record.S3.Object.Key,
					record.S3.Object.Size,
				})

				for scanner.Scan() {
					cap.SetData([]byte(scanner.Text()))
					sub.SendTransform(cap)
				}
			}
		}

		sub.TransformSignal()
		transformWg.Wait()
		sub.SinkSignal()
		sinkWg.Wait()
	}()

	if err := sub.Block(ctx); err != nil {
		return fmt.Errorf("sns-s3 handler: %v", err)
	}

	return nil
}

type snsMetadata struct {
	EventSubscriptionArn string    `json:"event_subscription_arn"`
	MessageID            string    `json:"message_id"`
	Subject              string    `json:"subject"`
	Timestamp            time.Time `json:"timestamp"`
}

func snsHandler(ctx context.Context, event events.SNSEvent) error {
	conf, err := appconfig.GetPrefetch(ctx)
	if err != nil {
		return fmt.Errorf("sns handler: %v", err)
	}
	json.Unmarshal(conf, &sub.Config)

	sub.CreateChannels(concurrency)
	defer sub.KillSignal()

	go func() {
		var sinkWg sync.WaitGroup
		var transformWg sync.WaitGroup

		sinkWg.Add(1)
		go sub.Sink(ctx, &sinkWg)

		for w := 0; w <= concurrency; w++ {
			transformWg.Add(1)
			go sub.Transform(ctx, &transformWg)
		}

		for _, record := range event.Records {
			cap := config.NewCapsule()
			cap.SetMetadata(snsMetadata{
				record.EventSubscriptionArn,
				record.SNS.MessageID,
				record.SNS.Subject,
				record.SNS.Timestamp,
			})

			cap.SetData([]byte(record.SNS.Message))
			sub.SendTransform(cap)
		}

		sub.TransformSignal()
		transformWg.Wait()
		sub.SinkSignal()
		sinkWg.Wait()
	}()

	if err := sub.Block(ctx); err != nil {
		return fmt.Errorf("sns handler: %v", err)
	}

	return nil
}

type sqsMetadata struct {
	EventSourceArn string            `json:"event_source_arn"`
	MessageID      string            `json:"message_id"`
	BodyMd5        string            `json:"body_md5"`
	Attributes     map[string]string `json:"attributes"`
}

func sqsHandler(ctx context.Context, event events.SQSEvent) error {
	conf, err := appconfig.GetPrefetch(ctx)
	if err != nil {
		return fmt.Errorf("sqs handler: %v", err)
	}
	json.Unmarshal(conf, &sub.Config)

	sub.CreateChannels(concurrency)
	defer sub.KillSignal()

	go func() {
		var sinkWg sync.WaitGroup
		var transformWg sync.WaitGroup

		sinkWg.Add(1)
		go sub.Sink(ctx, &sinkWg)

		for w := 0; w <= concurrency; w++ {
			transformWg.Add(1)
			go sub.Transform(ctx, &transformWg)
		}

		for _, msg := range event.Records {
			cap := config.NewCapsule()
			cap.SetData([]byte(msg.Body))
			cap.SetMetadata(sqsMetadata{
				msg.EventSourceARN,
				msg.MessageId,
				msg.Md5OfBody,
				msg.Attributes,
			})
			sub.SendTransform(cap)
		}

		sub.TransformSignal()
		transformWg.Wait()
		sub.SinkSignal()
		sinkWg.Wait()
	}()

	if err := sub.Block(ctx); err != nil {
		return fmt.Errorf("sns handler: %v", err)
	}

	return nil
}
