package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/brexhq/substation"
	"github.com/brexhq/substation/internal/file"
)

var (
	handler string

	// errLambdaMissingHandler is returned when the Lambda is deployed without a configured handler.
	errLambdaMissingHandler = fmt.Errorf("SUBSTATION_LAMBDA_HANDLER environment variable is missing")

	// errLambdaInvalidHandler is returned when the Lambda is deployed with an unsupported handler.
	errLambdaInvalidHandler = fmt.Errorf("SUBSTATION_LAMBDA_HANDLER environment variable is invalid")

	// errLambdaInvalidJSON is returned when the Lambda is deployed with a transform that produces invalid JSON.
	errLambdaInvalidJSON = fmt.Errorf("transformed data is invalid JSON and cannot be returned")
)

type customConfig struct {
	substation.Config

	Concurrency int `json:"concurrency"`
}

// getConfig contextually retrieves a Substation configuration.
func getConfig(ctx context.Context) (io.Reader, error) {
	buf := new(bytes.Buffer)

	cfg, found := os.LookupEnv("SUBSTATION_CONFIG")
	if !found {
		return nil, fmt.Errorf("no config found")
	}

	path, err := file.Get(ctx, cfg)
	defer os.Remove(path)

	if err != nil {
		return nil, err
	}

	conf, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer conf.Close()

	if _, err := io.Copy(buf, conf); err != nil {
		return nil, err
	}

	return buf, nil
}

func main() {
	switch h := handler; h {
	case "AWS_API_GATEWAY":
		lambda.Start(gatewayHandler)
	case "AWS_DYNAMODB_STREAM", "AWS_DYNAMODB": // AWS_DYNAMODB is deprecated
		lambda.Start(dynamodbHandler)
	case "AWS_KINESIS_DATA_FIREHOSE":
		lambda.Start(firehoseHandler)
	case "AWS_KINESIS_DATA_STREAM", "AWS_KINESIS": // AWS_KINESIS is deprecated
		lambda.Start(kinesisStreamHandler)
	case "AWS_LAMBDA":
		lambda.Start(lambdaHandler)
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
	var ok bool
	handler, ok = os.LookupEnv("SUBSTATION_LAMBDA_HANDLER")
	if !ok {
		panic(fmt.Errorf("init handler %s: %v", handler, errLambdaMissingHandler))
	}
}
