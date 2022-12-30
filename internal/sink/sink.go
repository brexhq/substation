package sink

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
)

// errInvalidFactoryInput is returned when an unsupported Sink is referenced in Factory.
const errInvalidFactoryInput = errors.Error("invalid factory input")

type sink interface {
	Send(context.Context, *config.Channel) error
}

// Make returns a configured sink from a sink configuration.
func Make(cfg config.Config) (sink, error) {
	switch t := cfg.Type; t {
	case "aws_dynamodb":
		var s _awsDynamodb
		_ = config.Decode(cfg.Settings, &s)
		return &s, nil
	case "aws_kinesis":
		var s _awsKinesis
		_ = config.Decode(cfg.Settings, &s)
		return &s, nil
	case "aws_kinesis_firehose":
		var s _awsKinesisFirehose
		_ = config.Decode(cfg.Settings, &s)
		return &s, nil
	case "aws_s3":
		var s _awsS3
		_ = config.Decode(cfg.Settings, &s)
		return &s, nil
	case "aws_sqs":
		var s _awsSQS
		_ = config.Decode(cfg.Settings, &s)
		return &s, nil
	case "grpc":
		var s _grpc
		_ = config.Decode(cfg.Settings, &s)
		return &s, nil
	case "http":
		var s _http
		_ = config.Decode(cfg.Settings, &s)
		return &s, nil
	case "stdout":
		var s _stdout
		_ = config.Decode(cfg.Settings, &s)
		return &s, nil
	case "sumologic":
		var s _sumologic
		_ = config.Decode(cfg.Settings, &s)
		return &s, nil
	default:
		return nil, fmt.Errorf("sink: settings %v: %v", cfg.Settings, errInvalidFactoryInput)
	}
}
