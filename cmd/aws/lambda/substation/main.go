package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/brexhq/substation"
	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/file"
)

var (
	handler      string
	functionName string
)

// errLambdaMissingHandler is returned when the Lambda is deployed without a configured handler.
var errLambdaMissingHandler = fmt.Errorf("missing SUBSTATION_HANDLER environment variable")

// errLambdaInvalidHandler is returned when the Lambda is deployed with an unsupported handler.
var errLambdaInvalidHandler = fmt.Errorf("invalid handler")

type customConfig struct {
	substation.Config

	Concurrency int           `json:"concurrency"`
	Metrics     config.Config `json:"metrics"`
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
	case "AWS_API_GATEWAY_ASYNC", "AWS_API_GATEWAY": // AWS_API_GATEWAY is deprecated
		lambda.Start(gatewayHandler)
	case "AWS_DYNAMODB_STREAM", "AWS_DYNAMODB": // AWS_DYNAMODB is deprecated
		lambda.Start(dynamodbHandler)
	case "AWS_KINESIS_DATA_FIREHOSE":
		lambda.Start(firehoseHandler)
	case "AWS_KINESIS_DATA_STREAM", "AWS_KINESIS": // AWS_KINESIS is deprecated
		lambda.Start(kinesisHandler)
	case "AWS_LAMBDA_ASYNC":
		lambda.Start(lambdaAsyncHandler)
	case "AWS_LAMBDA_SYNC":
		lambda.Start(lambdaSyncHandler)
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
	handler, ok = os.LookupEnv("SUBSTATION_HANDLER")
	if !ok {
		panic(fmt.Errorf("init handler %s: %v", handler, errLambdaMissingHandler))
	}

	functionName, _ = os.LookupEnv("AWS_LAMBDA_FUNCTION_NAME")
}
