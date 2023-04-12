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
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/log"
)

const (
	// errAWSS3EmptyPrefix is returned when the sink is configured with a prefix
	// key, but the key is not found in the object or the key is empty.
	errAWSS3EmptyPrefix = errors.Error("empty prefix string")
	// errAWSS3EmptySuffix is returned when the sink is configured with a suffix
	// key, but the key is not found in the object or the key is empty.
	errAWSS3EmptySuffix = errors.Error("empty suffix string")
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
	// the prefix prepended to the object path. If used, then
	// this overrides Prefix.
	//
	// This is optional and has no default.
	PrefixKey string `json:"prefix_key"`
	// FilePath determines how the name of the uploaded object is constructed.
	// One of these formats is constructed depending on the configuration:
	//
	// - prefix/date_format/uuid.extension
	//
	// - prefix/date_format/uuid/suffix.extension
	FilePath filePath `json:"file_path"`
}

// Send sinks a channel of encapsulated data with the sink.
func (s *sinkAWSS3) Send(ctx context.Context, ch *config.Channel) error {
	if !s3uploader.IsEnabled() {
		s3uploader.Setup()
	}

	files := make(map[string]*os.File)

	object := s.FilePath.New()
	if object == "" {
		// default object name is:
		// - year, month, and day
		// - random UUID
		object = time.Now().Format("2006/01/02") + "/" + uuid.New().String()
		// currently only supports gzip compression
		object += ".gz"

		// TODO: remove in v1.0.0
		if s.Prefix != "" {
			object = s.Prefix + "/" + object
		}
	}

	// newline character for Unix-based systems
	separator := []byte("\n")

	for capsule := range ch.C {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// innerObject is used so that key values can be interpolated into the object name
			innerObject := object

			// if either prefix or suffix keys are set, then the object name is non-default
			// and can be safely interpolated. if either are empty strings, then an error
			// is returned.
			if s.FilePath.PrefixKey != "" {
				prefix := capsule.Get(s.FilePath.PrefixKey).String()
				if prefix == "" {
					return fmt.Errorf("sink: aws_s3: bucket %s object %s: %v", s.Bucket, object, errAWSS3EmptyPrefix)
				}

				innerObject = strings.Replace(innerObject, "${PATH_PREFIX}", prefix, 1)
			}
			if s.FilePath.SuffixKey != "" {
				suffix := capsule.Get(s.FilePath.SuffixKey).String()
				if suffix == "" {
					return fmt.Errorf("sink: aws_s3: bucket %s object %s: %v", s.Bucket, object, errAWSS3EmptySuffix)
				}

				innerObject = strings.Replace(innerObject, "${PATH_SUFFIX}", suffix, 1)
			}

			// TODO: remove in v1.0.0
			if s.PrefixKey != "" {
				prefix := capsule.Get(s.FilePath.PrefixKey).String()
				innerObject = prefix + "/" + object
			}

			if _, ok := files[innerObject]; !ok {
				f, err := os.CreateTemp("", "substation")
				if err != nil {
					return fmt.Errorf("sink: aws_s3: bucket %s object %s: %v", s.Bucket, innerObject, err)
				}

				defer os.Remove(f.Name()) //nolint:staticcheck // SA9001: channel is closed on error, defer will run
				defer f.Close()           //nolint:staticcheck // SA9001: channel is closed on error, defer will run

				files[innerObject] = f
			}

			if _, err := files[innerObject].Write(capsule.Data()); err != nil {
				return fmt.Errorf("sink: aws_s3: bucket %s object %s: %v", s.Bucket, innerObject, err)
			}
			if _, err := files[innerObject].Write(separator); err != nil {
				return fmt.Errorf("sink: aws_s3: bucket %s object %s: %v", s.Bucket, innerObject, err)
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
	for object, file := range files {
		if _, err := file.Seek(0, 0); err != nil {
			return fmt.Errorf("sink: aws_s3: bucket %s: %v", s.Bucket, err)
		}

		reader, writer := io.Pipe()
		defer reader.Close()

		// goroutine avoids deadlock
		go func() {
			// currently only supports gzip compression
			gz := gzip.NewWriter(writer)
			defer writer.Close()
			defer gz.Close()

			_, _ = io.Copy(gz, file)
		}()

		if _, err := s3uploader.Upload(ctx, s.Bucket, object, reader); err != nil {
			return fmt.Errorf("sink: aws_s3: bucket %s object %s: %v", s.Bucket, object, err)
		}

		// s3uploader.Upload does not return the size of uploaded data, so we use the size of the uncompressed file when reporting stats for debugging
		fs, err := file.Stat()
		if err != nil {
			return fmt.Errorf("sink: aws_s3: bucket %s object %s: %v", s.Bucket, object, err)
		}

		log.WithField(
			"bucket", s.Bucket,
		).WithField(
			"object", object,
		).WithField(
			"size", fs.Size(),
		).Debug("uploaded data to S3")
	}

	return nil
}
