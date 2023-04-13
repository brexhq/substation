package sink

import (
	"compress/gzip"
	"compress/zlib"
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
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

type fw struct {
	*os.File
	ft  string
	sep []byte

	gz *gzip.Writer
	zl *zlib.Writer
}

func (f *fw) Type() string {
	return f.ft
}

func (f *fw) Write(b []byte) (int, error) {
	if strings.HasPrefix(f.ft, "text") {
		b = append(b, f.sep...)
	}

	switch f.ft {
	case "gzip":
		fallthrough
	case "text_gzip":
		return f.gz.Write(b)
	case "zlib":
		fallthrough
	case "text_zlib":
		return f.gz.Write(b)
	case "text":
		fallthrough
	default:
		return f.File.Write(b)
	}
}

func (f *fw) Close() error {
	switch f.ft {
	case "gzip":
		fallthrough
	case "text_gzip":
		if err := f.gz.Close(); err != nil {
			return err
		}
	case "zlib":
		fallthrough
	case "text_zlib":
		if err := f.zl.Close(); err != nil {
			return err
		}
	}

	if err := f.File.Close(); err != nil {
		return err
	}

	return nil
}

func NewFileWrapper(f *os.File, fmt config.Config) *fw {
	var sep []byte
	switch runtime.GOOS {
	case "windows":
		sep = []byte("\r\n")
	default:
		sep = []byte("\n")
	}

	switch fmt.Type {
	case "gzip":
		fallthrough
	case "text_gzip":
		return &fw{f, fmt.Type, sep, gzip.NewWriter(f), nil}
	case "zlib":
		fallthrough
	case "text_zlib":
		return &fw{f, fmt.Type, sep, nil, zlib.NewWriter(f)}
	case "text":
		return &fw{f, fmt.Type, sep, nil, nil}
	default:
		return &fw{f, "data", sep, nil, nil}
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
	// This is optional and defaults to false.
	UUID bool `json:"uuid"`
	// Extension appends a file extension to the filename.
	//
	// This is optional and has no default.
	Extension string `json:"extension"`
}

// New constructs a file path using the pattern
// [prefix]/[time_format]/[uuid]/[suffix][.extension], where each field is optional
// and builds on the previous field.
//
// If only one field is set, then this constructs a filename,
// otherwise it constructs a file path.
//
// If the struct is empty, then this returns an empty string. The caller is
// responsible for creating a default file path if needed.
func (p filePath) New() string {
	// if all of these fields are empty, then a valid file path cannot be constructed.
	if p.Prefix == "" && p.PrefixKey == "" &&
		p.Suffix == "" && p.SuffixKey == "" &&
		p.TimeFormat == "" && !p.UUID {
		return ""
	}

	// temporarily storing values for the file path in an array allows for any
	// individual field to be used as the filename if no other fields are set.
	arr := []string{}

	// PrefixKey takes precedence over Prefix.
	switch {
	case p.PrefixKey != "":
		arr = append(arr, "${PATH_PREFIX}")
	case p.Prefix != "":
		arr = append(arr, p.Prefix)
	}

	if p.TimeFormat != "" {
		now := time.Now()

		switch p.TimeFormat {
		case "unix":
			arr = append(arr, fmt.Sprintf("%d", now.Unix()))
		case "unix_milli":
			arr = append(arr, fmt.Sprintf("%d", now.UnixMilli()))
		default:
			arr = append(arr, now.Format(p.TimeFormat))
		}
	}

	if p.UUID {
		arr = append(arr, uuid.NewString())
	}

	// SuffixKey takes precedence over Suffix.
	switch {
	case p.SuffixKey != "":
		arr = append(arr, "${PATH_SUFFIX}")
	case p.Suffix != "":
		arr = append(arr, p.Suffix)
	}

	// if only one field is set, then this is only a filename, otherwise
	// it is a file path.
	path := strings.Join(arr, "/")

	// this works regardless of whether an extension is set.
	return path + p.Extension
}
