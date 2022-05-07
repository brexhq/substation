package sink

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
)

// SinkInvalidFactoryConfig is used when an unsupported Sink is referenced in Factory
const SinkInvalidFactoryConfig = errors.Error("SinkInvalidFactoryConfig")

// Sink is the interface used by all Substation sinks. Sinks read channels of bytes and are interruptable via an anonymous struct channel.
type Sink interface {
	Send(context.Context, chan []byte, chan struct{}) error
}

// Factory loads Sinks from a Config. This is the recommended function for retrieving ready-to-use Sinks.
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
		return nil, fmt.Errorf("err retrieving %s from factory: %v", t, SinkInvalidFactoryConfig)
	}
}

// randomString returns a randomly generated string. This is referenced from https://golangdocs.com/generate-random-string-in-golang.
func randomString() string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	s := make([]rune, 16)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}
