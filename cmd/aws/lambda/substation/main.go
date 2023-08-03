package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/brexhq/substation/internal/file"
)

var handler string

// errLambdaMissingHandler is returned when the Lambda is deployed without a configured handler.
var errLambdaMissingHandler = fmt.Errorf("missing SUBSTATION_HANDLER environment variable")

// errLambdaInvalidHandler is returned when the Lambda is deployed with an unsupported handler.
var errLambdaInvalidHandler = fmt.Errorf("invalid handler")

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
	case "AWS_DYNAMODB":
		lambda.Start(dynamodbHandler)
	case "AWS_KINESIS":
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
	var found bool
	handler, found = os.LookupEnv("SUBSTATION_HANDLER")
	if !found {
		panic(fmt.Errorf("init handler %s: %v", handler, errLambdaMissingHandler))
	}
}
