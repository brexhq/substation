package sink

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/aws/s3manager"
	"github.com/brexhq/substation/internal/log"
)

var s3uploader s3manager.UploaderAPI

// awsS3 sinks data as objects to an AWS S3 bucket.
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
	// See filePath.New for more information.
	FilePath filePath `json:"file_path"`
	// FileFormat determines the format of the file. These file formats are
	// supported:
	//
	// - data (binary data)
	//
	// - json
	//
	// - text
	//
	// If the format type does not have a common file extension, then
	// no extension is added to the file name.
	//
	// Defaults to json.
	FileFormat config.Config `json:"file_format"`
	// FileCompression determines the compression type applied to the file.
	// These compression types are supported:
	//
	// - gzip (https://en.wikipedia.org/wiki/Gzip)
	//
	// - snappy (https://en.wikipedia.org/wiki/Snappy_(compression))
	//
	// - zstd (https://en.wikipedia.org/wiki/Zstd)
	//
	// If the compression type does not have a common file extension, then
	// no extension is added to the file name.
	//
	// Defaults to gzip.
	FileCompression config.Config `json:"file_compression"`
}

// Send sinks a channel of encapsulated data with the sink.
//
//nolint:gocognit
func (s *sinkAWSS3) Send(ctx context.Context, ch *config.Channel) error {
	if !s3uploader.IsEnabled() {
		s3uploader.Setup()
	}

	files := make(map[string]*fw)

	// file extensions are dynamic and not directly configurable
	extension := NewFileExtension(s.FileFormat, s.FileCompression)
	now := time.Now()

	// default object key is: year/month/day/uuid.extension
	object := s.FilePath.New()
	if object == "" {
		object = path.Join(
			now.Format("2006"), now.Format("01"), now.Format("02"),
			uuid.New().String(),
		) + extension
	} else if s.FilePath.Extension {
		object += extension
	}

	// provides backward compatibility for v0.8.4
	// TODO(v1.0.0): remove this
	if s.FileCompression.Type == "" && s.FileFormat.Type == "" &&
		s.FilePath.New() == "" {
		// TODO: move to constructor
		if s.FileFormat.Type == "" {
			s.FileFormat.Type = "json"
		}

		// TODO: move to constructor
		if s.FileCompression.Type == "" {
			s.FileCompression.Type = "gzip"
		}

		object = path.Join(
			// path.Join ignores empty strings, so this is safe
			s.Prefix,
			now.Format("2006"), now.Format("01"), now.Format("02"),
			uuid.New().String(),
		) + ".gz"
	}

	for capsule := range ch.C {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// innerObject is used so that key values can be interpolated into the object key
			innerObject := object

			// if either prefix or suffix keys are set, then the object key is non-default
			// and can be safely interpolated. if either are empty strings, then an error
			// is returned.
			if s.FilePath.PrefixKey != "" {
				prefix := capsule.Get(s.FilePath.PrefixKey).String()
				if prefix == "" {
					return fmt.Errorf("sink: aws_s3: bucket %s object %s: %v", s.Bucket, object, errEmptyPrefix)
				}

				innerObject = strings.Replace(innerObject, "${PATH_PREFIX}", prefix, 1)
			}
			if s.FilePath.SuffixKey != "" {
				suffix := capsule.Get(s.FilePath.SuffixKey).String()
				if suffix == "" {
					return fmt.Errorf("sink: aws_s3: bucket %s object %s: %v", s.Bucket, object, errEmptySuffix)
				}

				innerObject = strings.Replace(innerObject, "${PATH_SUFFIX}", suffix, 1)
			}

			// TODO(v1.0.0): remove this
			if s.PrefixKey != "" && s.FilePath.New() == "" {
				prefix := capsule.Get(s.PrefixKey).String()
				innerObject = path.Join(prefix, object)
			}

			if _, ok := files[innerObject]; !ok {
				f, err := os.CreateTemp("", "substation")
				if err != nil {
					return fmt.Errorf("sink: aws_s3: bucket %s object %s: %v", s.Bucket, innerObject, err)
				}

				defer os.Remove(f.Name()) //nolint:staticcheck // SA9001: channel is closed on error, defer will run

				if files[innerObject], err = NewFileWrapper(f, s.FileFormat, s.FileCompression); err != nil {
					return fmt.Errorf("sink: file: file_path %s: %v", innerObject, err)
				}

				defer files[innerObject].Close() //nolint:staticcheck // SA9001: channel is closed on error, defer will run
			}

			if _, err := files[innerObject].Write(capsule.Data()); err != nil {
				return fmt.Errorf("sink: aws_s3: bucket %s object %s: %v", s.Bucket, innerObject, err)
			}
		}
	}

	for object, file := range files {
		// close to flush the file buffers before uploading to S3
		if err := file.Close(); err != nil {
			return fmt.Errorf("sink: aws_s3: bucket %s object %s: %v", s.Bucket, object, err)
		}

		// s3uploader requires an open file (reader)
		f, err := os.Open(file.Name())
		if err != nil {
			return fmt.Errorf("sink: aws_s3: bucket %s object %s: %v", s.Bucket, object, err)
		}
		defer f.Close()

		if _, err := s3uploader.Upload(ctx, s.Bucket, object, f); err != nil {
			return fmt.Errorf("sink: aws_s3: bucket %s object %s: %v", s.Bucket, object, err)
		}

		// s3uploader.Upload does not return the size of uploaded data, so we use the size of
		// the  file when reporting stats for debugging
		fs, err := f.Stat()
		if err != nil {
			return fmt.Errorf("sink: aws_s3: bucket %s object %s: %v", s.Bucket, object, err)
		}

		log.WithField(
			"bucket", s.Bucket,
		).WithField(
			"object", object,
		).WithField(
			"size", fs.Size(),
		).WithField(
			"format", s.FileFormat.Type,
		).WithField(
			"compression", s.FileCompression.Type,
		).Debug("uploaded data to S3")
	}

	return nil
}
