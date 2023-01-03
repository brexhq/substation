package sink

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
)

// errInvalidFactoryInput is returned when an unsupported Sink is referenced in Factory.
const errInvalidFactoryInput = errors.Error("invalid factory input")

type Sink interface {
	Send(context.Context, *config.Channel) error
}

// New returns a configured Sink from a sink configuration.
func New(cfg config.Config) (Sink, error) {
	switch t := cfg.Type; t {
	case "aws_dynamodb":
		var s sinkAWSDynamoDB
		_ = config.Decode(cfg.Settings, &s)
		return &s, nil
	case "aws_kinesis":
		var s sinkAWSKinesis
		_ = config.Decode(cfg.Settings, &s)
		return &s, nil
	case "aws_kinesis_firehose":
		var s sinkAWSKinesisFirehose
		_ = config.Decode(cfg.Settings, &s)
		return &s, nil
	case "aws_s3":
		var s sinkAWSS3
		_ = config.Decode(cfg.Settings, &s)
		return &s, nil
	case "aws_sqs":
		var s sinkAWSSQS
		_ = config.Decode(cfg.Settings, &s)
		return &s, nil
	case "grpc":
		var s sinkGRPC
		_ = config.Decode(cfg.Settings, &s)
		return &s, nil
	case "http":
		var s sinkHTTP
		_ = config.Decode(cfg.Settings, &s)
		return &s, nil
	case "stdout":
		var s sinkStdout
		_ = config.Decode(cfg.Settings, &s)
		return &s, nil
	case "sumologic":
		var s sinkSumoLogic
		_ = config.Decode(cfg.Settings, &s)
		return &s, nil
	default:
		return nil, fmt.Errorf("Sink: settings %v: %v", cfg.Settings, errInvalidFactoryInput)
	}
}
