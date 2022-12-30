package sink

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
)

// errInvalidFactoryInput is returned when an unsupported Sink is referenced in Factory.
const errInvalidFactoryInput = errors.Error("invalid factory input")

// Sink is an interface for sending data to external services. Sinks read channels of capsules and are interruptable.
type Sink interface {
	Send(context.Context, *config.Channel) error
}

// Factory returns a configured Sink from a config. This is the recommended method for retrieving ready-to-use Sinks.
func Factory(cfg config.Config) (Sink, error) {
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
		var s Grpc
		_ = config.Decode(cfg.Settings, &s)
		return &s, nil
	case "http":
		var s HTTP
		_ = config.Decode(cfg.Settings, &s)
		return &s, nil
	case "stdout":
		var s Stdout
		_ = config.Decode(cfg.Settings, &s)
		return &s, nil
	case "sumologic":
		var s SumoLogic
		_ = config.Decode(cfg.Settings, &s)
		return &s, nil
	default:
		return nil, fmt.Errorf("sink: settings %v: %v", cfg.Settings, errInvalidFactoryInput)
	}
}
