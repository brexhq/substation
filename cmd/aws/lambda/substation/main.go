package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"sync"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/kinesis-aggregation/go/deaggregator"

	"github.com/brexhq/substation/cmd"
	"github.com/brexhq/substation/internal/aws/appconfig"
	"github.com/brexhq/substation/internal/aws/kinesis"
	"github.com/brexhq/substation/internal/aws/s3manager"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/json"
	"github.com/brexhq/substation/internal/log"
)

var sub cmd.Substation
var handler string

// LambdaMissingHandler is used when the Lambda is deployed without a configured handler
const LambdaMissingHandler = errors.Error("LambdaMissingHandler")

// LambdaUnsupportedHandler is used when the Lambda is deployed without a supported handler
const LambdaUnsupportedHandler = errors.Error("LambdaUnsupportedHandler")

func main() {
	switch h := handler; h {
	case "GATEWAY":
		lambda.Start(gateway)
	case "KINESIS":
		lambda.Start(kinesisHandler)
	case "S3":
		lambda.Start(s3)
	case "SNS":
		lambda.Start(sns)
	default:
		panic(fmt.Errorf("err handler %s: %v", h, LambdaUnsupportedHandler))
	}
}

func init() {
	var found bool
	handler, found = os.LookupEnv("SUBSTATION_HANDLER")
	if !found {
		panic(fmt.Errorf("err handler %s: %v", handler, LambdaMissingHandler))
	}
}

func gateway(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	conf, err := appconfig.GetPrefetch(ctx)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500}, fmt.Errorf("err failed to fetch AppConfig configuration: %v", err)
	}
	json.Unmarshal(conf, &sub.Config)

	sub.CreateChannels(runtime.NumCPU())
	defer sub.KillSignal()

	go func() {
		var sinkWg sync.WaitGroup
		var transformWg sync.WaitGroup

		sinkWg.Add(1)
		go sub.Sink(ctx, &sinkWg)

		transformWg.Add(1)
		go sub.Transform(ctx, &transformWg)

		if len(request.Body) != 0 {
			sub.SendTransform([]byte(request.Body))
		}

		sub.TransformSignal()
		transformWg.Wait()
		sub.SinkSignal()
		sinkWg.Wait()
	}()

	if err := sub.Block(ctx); err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500}, err
	}

	return events.APIGatewayProxyResponse{StatusCode: 200}, nil
}

func kinesisHandler(ctx context.Context, kinesisEvent events.KinesisEvent) error {
	conf, err := appconfig.GetPrefetch(ctx)
	if err != nil {
		return fmt.Errorf("err failed to fetch AppConfig configuration: %v", err)
	}
	json.Unmarshal(conf, &sub.Config)

	sub.CreateChannels(runtime.NumCPU())
	defer sub.KillSignal()

	go func() {
		var sinkWg sync.WaitGroup
		var transformWg sync.WaitGroup

		sinkWg.Add(1)
		go sub.Sink(ctx, &sinkWg)

		for w := 0; w <= runtime.NumCPU(); w++ {
			transformWg.Add(1)
			go sub.Transform(ctx, &transformWg)
		}

		converted := kinesis.ConvertEventsRecords(kinesisEvent.Records)
		deaggregated, err := deaggregator.DeaggregateRecords(converted)
		if err != nil {
			sub.SendErr(fmt.Errorf("err failed to deaggregate Kinesis records: %v", err))
			return
		}

		for _, record := range deaggregated {
			if len(record.Data) == 0 {
				continue
			}

			sub.SendTransform(record.Data)
		}

		sub.TransformSignal()
		transformWg.Wait()
		sub.SinkSignal()
		sinkWg.Wait()
	}()

	if err := sub.Block(ctx); err != nil {
		return err
	}

	return nil
}

func s3(ctx context.Context, s3Event events.S3Event) error {
	conf, err := appconfig.GetPrefetch(ctx)
	if err != nil {
		return fmt.Errorf("err failed to fetch AppConfig configuration: %v", err)
	}
	json.Unmarshal(conf, &sub.Config)

	sub.CreateChannels(runtime.NumCPU())
	defer sub.KillSignal()

	go func() {
		api := s3manager.DownloaderAPI{}
		api.Setup()

		var sinkWg sync.WaitGroup
		var transformWg sync.WaitGroup

		sinkWg.Add(1)
		go sub.Sink(ctx, &sinkWg)

		for w := 0; w <= runtime.NumCPU(); w++ {
			transformWg.Add(1)
			go sub.Transform(ctx, &transformWg)
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
				sub.SendErr(fmt.Errorf("err failed to download bucket %s key %s as scanner: %v", record.S3.Bucket.Name, record.S3.Object.Key, err))
				return
			}

			for scanner.Scan() {
				scanData := scanner.Bytes()
				if len(scanData) == 0 {
					continue
				}
				sub.SendTransform(scanData)
			}
		}

		sub.TransformSignal()
		transformWg.Wait()
		sub.SinkSignal()
		sinkWg.Wait()
	}()

	if err := sub.Block(ctx); err != nil {
		return err
	}

	return nil
}

func sns(ctx context.Context, snsEvent events.SNSEvent) error {
	conf, err := appconfig.GetPrefetch(ctx)
	if err != nil {
		return fmt.Errorf("err failed to fetch AppConfig configuration: %v", err)
	}
	json.Unmarshal(conf, &sub.Config)

	sub.CreateChannels(runtime.NumCPU())
	defer sub.KillSignal()

	go func() {
		// SNS pulls data from S3
		api := s3manager.DownloaderAPI{}
		api.Setup()

		var sinkWg sync.WaitGroup
		var transformWg sync.WaitGroup

		sinkWg.Add(1)
		go sub.Sink(ctx, &sinkWg)

		for w := 0; w <= runtime.NumCPU(); w++ {
			transformWg.Add(1)
			go sub.Transform(ctx, &transformWg)
		}

		for _, record := range snsEvent.Records {
			var s3Event events.S3Event
			err := json.Unmarshal([]byte(record.SNS.Message), &s3Event)
			if err != nil {
				sub.SendErr(fmt.Errorf("err failed to unmarshal SNS message as S3Event: %v", err))
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
					sub.SendErr(fmt.Errorf("err failed to download bucket %s key %s as scanner: %v", record.S3.Bucket.Name, record.S3.Object.Key, err))
					return
				}

				for scanner.Scan() {
					scanData := scanner.Bytes()
					if len(scanData) == 0 {
						continue
					}
					sub.SendTransform(scanData)
				}
			}
		}

		sub.TransformSignal()
		transformWg.Wait()
		sub.SinkSignal()
		sinkWg.Wait()
	}()

	if err := sub.Block(ctx); err != nil {
		return err
	}

	return nil
}
