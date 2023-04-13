package sink

import (
	"context"
	"fmt"
	"time"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/google/uuid"
)

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
	case "file":
		var s sinkFile
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
		return nil, fmt.Errorf("sink: settings %v: %v", cfg.Settings, errors.ErrInvalidFactoryInput)
	}
}

type filePath struct {
	// Prefix is a prefix prepended to the file path.
	//
	// This is optional and has no default.
	Prefix string `json:"prefix"`
	// PrefixKey retrieves a value from an object that is used as
	// the prefix prepended to the file path. If used, then
	// this overrides Prefix.
	//
	// This is optional and has no default.
	PrefixKey string `json:"prefix_key"`
	// Suffix is a suffix appended to the file path and is used as
	// the object filename.
	//
	// This is optional and has no default.
	Suffix string `json:"suffix"`
	// SuffixKey retrieves a value from an object that is used as
	// the suffix appended to the file path. If used, then
	// this overrides Suffix.
	//
	// This is optional and has no default.
	SuffixKey string `json:"suffix_key"`
	// TimeFormat inserts a formatted datetime string into the file path.
	// The string uses pattern-based layouts
	// (https://gobyexample.com/procTime-formatting-parsing).
	//
	// This is optional and has no default.
	TimeFormat string `json:"time_format"`
	// UUID inserts a random UUID into the file path. If a suffix is
	// not set, then this is used as the filename.
	//
	// This is optional and has no default.
	UUID *bool `json:"uuid"`
	// Extension appends a file extension to the filename.
	//
	// This is optional and has no default.
	Extension *bool `json:"extension"`
}

// New constructs a file path that follows one of these formats depending on the configuration:
//
// - [prefix]/[time_format]/[uuid].[extension]
//
// - [prefix]/[time_format]/[uuid]/[suffix].[extension]
//
// If the struct is empty, then this returns an empty string.
func (p filePath) New() (path string) {
	// PrefixKey takes precedence over Prefix.
	switch {
	case p.PrefixKey != "":
		path = "${PATH_PREFIX}/"
	case p.Prefix != "":
		path = p.Prefix + "/"
	}

	if p.TimeFormat != "" {
		now := time.Now()

		switch p.TimeFormat {
		case "unix":
			path += fmt.Sprintf("%d/", now.Unix())
		case "unix_milli":
			path += fmt.Sprintf("%d/", now.UnixMilli())
		default:
			path += now.Format(p.TimeFormat) + "/"
		}
	}

	// if suffix exists, then UUID is a directory and not a file. if it doesn't exist,
	// then UUID is a file.
	switch {
	case (p.Suffix != "" || p.SuffixKey != "") && p.UUID != nil && *p.UUID:
		path += uuid.NewString() + "/"
	case p.UUID != nil && *p.UUID:
		path += uuid.NewString()
	}

	// SuffixKey takes precedence over Suffix.
	switch {
	case p.SuffixKey != "":
		path += "${PATH_SUFFIX}"
	case p.Suffix != "":
		path += p.Suffix
	}

	if p.Extension == nil {
		return path
	}

	// if other file formats are supported, then this should be refactored
	// based on file type. the default should continue to be gzip.
	switch {
	default:
		path += ".gz"
	}

	return path
}
