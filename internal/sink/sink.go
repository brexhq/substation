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

const (
	// errEmptyPrefix is returned when a file-based sink is configured with a
	// prefix key, but the key is not found in the object or the key is empty.
	errEmptyPrefix = errors.Error("empty prefix string")
	// errEmptySuffix is returned when a file-based sink is configured with a
	// suffix key, but the key is not found in the object or the key is empty.
	errEmptySuffix = errors.Error("empty suffix string")
)

type Sink interface {
	Send(context.Context, *config.Channel) error
}

// New returns a configured Sink from a sink configuration.
func New(cfg config.Config) (Sink, error) {
	switch t := cfg.Type; t {
	case "aws_dynamodb":
		return newSinkAWSDynamoDB(cfg)
	case "aws_kinesis":
		return newSinkAWSKinesis(cfg)
	case "aws_kinesis_firehose":
		return newSinkAWSKinesisFirehose(cfg)
	case "aws_s3":
		return newSinkAWSS3(cfg)
	case "aws_sqs":
		return newSinkAWSSQS(cfg)
	case "file":
		var s sinkFile
		_ = config.Decode(cfg.Settings, &s)
		return &s, nil
	case "grpc":
		return newSinkGRPC(cfg)
	case "http":
		return newSinkHTTP(cfg)
	case "stdout":
		return newSinkStdout(cfg)
	case "sumologic":
		return newSinkSumoLogic(cfg)
	default:
		return nil, fmt.Errorf("sink: settings %v: %v", cfg.Settings, errors.ErrInvalidFactoryInput)
	}
}

type fw struct {
	*os.File
	w       io.WriteCloser
	newline []byte
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

	return f.File.Close()
}

func NewFileWrapper(f *os.File, format config.Config, compression config.Config) (*fw, error) {
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
	switch format.Type {
	case "data":
		newline = nil
	}

	switch compression.Type {
	case "gzip":
		return &fw{f, gzip.NewWriter(f), newline}, nil
	case "snappy":
		return &fw{f, snappy.NewBufferedWriter(f), newline}, nil
	case "zstd":
		// TODO: add settings support
		z, err := zstd.NewWriter(f)
		if err != nil {
			return nil, err
		}

		return &fw{f, z, newline}, nil
	default:
		return &fw{f, nil, newline}, nil
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
	// Must be one of:
	//
	// - pattern-based layouts (https://gobyexample.com/procTime-formatting-parsing)
	//
	// - unix: epoch (supports fractions of a second)
	//
	// - unix_milli: epoch milliseconds
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

		// these options mirror process/time.go
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
func NewFileExtension(format config.Config, compression config.Config) (ext string) {
	switch format.Type {
	case "data":
		break
	case "json":
		ext = ".json"
	case "text":
		ext = ".txt"
	}

	switch compression.Type {
	case "gzip":
		ext += ".gz"
	case "snappy":
		break
	case "zstd":
		ext += ".zst"
	}

	return ext
}
