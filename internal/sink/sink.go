package sink

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
)

// SinkInvalidFactoryConfig is returned when an unsupported Sink is referenced in Factory.
const SinkInvalidFactoryConfig = errors.Error("SinkInvalidFactoryConfig")

// Sink is an interface for sending data to external services. Sinks read channels of bytes and are interruptable via an anonymous struct channel.
type Sink interface {
	Send(context.Context, chan []byte, chan struct{}) error
}

// Factory loads a Sink from a Config. This is the recommended function for retrieving ready-to-use Sinks.
func Factory(cfg config.Config) (Sink, error) {
	switch t := cfg.Type; t {
	case "dynamodb":
		var s DynamoDB
		config.Decode(cfg.Settings, &s)
		return &s, nil
	case "http":
		var s HTTP
		config.Decode(cfg.Settings, &s)
		return &s, nil
	case "kinesis":
		var s Kinesis
		config.Decode(cfg.Settings, &s)
		return &s, nil
	case "s3":
		var s S3
		config.Decode(cfg.Settings, &s)
		return &s, nil
	case "stdout":
		var s Stdout
		config.Decode(cfg.Settings, &s)
		return &s, nil
	case "sumologic":
		var s SumoLogic
		config.Decode(cfg.Settings, &s)
		return &s, nil
	default:
		return nil, fmt.Errorf("sink settings %v: %v", cfg.Settings, SinkInvalidFactoryConfig)
	}
}
