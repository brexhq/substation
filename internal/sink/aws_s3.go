package sink

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/aws/s3manager"
	"github.com/brexhq/substation/internal/log"
)

var s3uploader s3manager.UploaderAPI

// awsS3 sinks data as gzip compressed objects to an AWS S3 bucket.
//
// Object names contain the year, month, and day the data was processed
// by the sink and can be optionally prefixed with a custom string.
type sinkAWSS3 struct {
	// Bucket is the AWS S3 bucket that data is written to.
	Bucket string `json:"bucket"`
	// Prefix is a prefix prepended to the object path.
	//
	// This is optional and has no default.
	Prefix string `json:"prefix"`
	// PrefixKey retrieves a value from an object that is used as
	// the prefix prepended to the S3 object path. If used, then
	// this overrides Prefix.
	//
	// This is optional and has no default.
	PrefixKey string `json:"prefix_key"`
}

// Send sinks a channel of encapsulated data with the sink.
func (s *sinkAWSS3) Send(ctx context.Context, ch *config.Channel) error {
	if !s3uploader.IsEnabled() {
		s3uploader.Setup()
	}

	files := make(map[string]*os.File)

	var prefix string
	if s.Prefix != "" {
		prefix = s.Prefix
	}

	// newline character for Unix-based systems
	separator := []byte("\n")

	for capsule := range ch.C {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if s.PrefixKey != "" {
				prefix = capsule.Get(s.PrefixKey).String()
			}

			if _, ok := files[prefix]; !ok {
				f, err := os.CreateTemp("", "substation")
				if err != nil {
					return fmt.Errorf("sink: aws_s3: bucket %s prefix %s: %v", s.Bucket, prefix, err)
				}

				defer os.Remove(f.Name()) //nolint:staticcheck // SA9001: channel is closed on error, defer will run
				defer f.Close()           //nolint:staticcheck // SA9001: channel is closed on error, defer will run

				files[prefix] = f
			}

			if _, err := files[prefix].Write(capsule.Data()); err != nil {
				return fmt.Errorf("sink: aws_s3: bucket %s prefix %s: %v", s.Bucket, prefix, err)
			}
			if _, err := files[prefix].Write(separator); err != nil {
				return fmt.Errorf("sink: aws_s3: bucket %s prefix %s: %v", s.Bucket, prefix, err)
			}
		}
	}

	/*
		uploading data to S3 requires the following steps:
			- reset offset to zero so the file content can be read
			- connect gzip writer to a pipe writer
			- copy file content to the gzip writer (copies to the pipe reader)
			- generate the S3 object key
			- upload the pipe reader to the bucket using the generated key
	*/
	for prefix, file := range files {
		if _, err := file.Seek(0, 0); err != nil {
			return fmt.Errorf("sink: aws_s3: bucket %s: %v", s.Bucket, err)
		}

		reader, w := io.Pipe()
		gz := gzip.NewWriter(w)

		// goroutine avoids deadlock
		go func() {
			defer w.Close()
			defer gz.Close()

			_, _ = io.Copy(gz, file)
		}()

		key := createKey(prefix)
		if _, err := s3uploader.Upload(ctx, s.Bucket, key, reader); err != nil {
			return fmt.Errorf("sink: aws_s3: bucket %s key %s: %v", s.Bucket, key, err)
		}

		// s3uploader.Upload does not return the size of uploaded data, so we use the size of the uncompressed file when reporting stats for debugging
		fs, err := file.Stat()
		if err != nil {
			return fmt.Errorf("sink: aws_s3: bucket %s key %s: %v", s.Bucket, key, err)
		}

		log.WithField(
			"bucket", s.Bucket,
		).WithField(
			"key", key,
		).WithField(
			"size", fs.Size(),
		).Debug("uploaded data to S3")
	}

	return nil
}

// createKey creates a date-based S3 object key that has this naming convention: [prefix : optional]/[year]/[month]/[day]/[uuid].gz
func createKey(prefix string) string {
	var key string

	if prefix != "" {
		key = prefix + "/"
	}

	now := time.Now().Format("2006-01-02")
	for _, date := range strings.Split(now, "-") {
		key += date + "/"
	}

	key = fmt.Sprint(key, uuid.NewString(), ".gz")
	return key
}
