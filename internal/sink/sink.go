package sink

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"time"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/google/uuid"
	"github.com/klauspost/compress/gzip"
	"github.com/klauspost/compress/snappy"
	"github.com/klauspost/compress/zstd"
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
	newline []byte

	w io.WriteCloser
}

func (f *fw) Write(b []byte) (int, error) {
	if f.newline != nil {
		b = append(b, f.newline...)
	}

	if f.w != nil {
		return f.w.Write(b)
	}

	return f.File.Write(b)
}

func (f *fw) Close() error {
	if f.w != nil {
		if err := f.w.Close(); err != nil {
			return err
		}
	}

	if err := f.File.Close(); err != nil {
		return err
	}

	return nil
}

func NewFileWrapper(f *os.File, fmt config.Config, cmp config.Config) (*fw, error) {
	var newline []byte
	switch runtime.GOOS {
	case "windows":
		newline = []byte("\r\n")
	default:
		newline = []byte("\n")
	}

	// if the file format is not text-based, then newline is unused.
	// if a file format uses a specific compression, then it should
	// be configured and returned in this switch.
	switch fmt.Type {
	case "data":
		newline = nil
	}

	switch cmp.Type {
	case "gzip":
		return &fw{f, newline, gzip.NewWriter(f)}, nil
	case "snappy":
		return &fw{f, newline, snappy.NewBufferedWriter(f)}, nil
	case "zstd":
		// TODO: add settings support
		z, err := zstd.NewWriter(f)
		if err != nil {
			return nil, err
		}

		return &fw{f, newline, z}, nil
	default:
		return &fw{f, newline, nil}, nil
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
	// This is optional and defaults to false.
	Extension bool `json:"extension"`
}

// New constructs a file path using the pattern
// [prefix]/[time_format]/[uuid]/[suffix], where each field is optional
// and builds on the previous field. The caller is responsible for
// creating an OS agnostic file path (filepath.FromSlash is recommended).
//
// If only one field is set, then this constructs a filename,
// otherwise it constructs a file path.
//
// If the struct is empty, then this returns an empty string. The caller is
// responsible for creating a default file path if needed.
func (p filePath) New() string {
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

	// if only one field is set, then this returns a filename, otherwise
	// it returns a file path.
	return path.Join(arr...)
}

// NewFileExtension returns a file extension based on file format and
// compression settings. The file extensions constructed by this function
// match this regular expression: `(\.json|\.txt)?(\.gz|\.zst)?`.
func NewFileExtension(fmt config.Config, cmp config.Config) (ext string) {
	switch fmt.Type {
	case "data":
		break
	case "json":
		ext = ".json"
	case "text":
		ext = ".txt"
	}

	switch cmp.Type {
	case "gzip":
		ext += ".gz"
	case "snappy":
		break
	case "zstd":
		ext += ".zst"
	}

	return ext
}
