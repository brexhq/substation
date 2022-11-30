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

/*
S3 sinks data as gzip compressed objects to an AWS S3 bucket. Object names contain the year, month, and day the data was processed by the sink; they can be optionally prefixed with a custom string.

The sink has these settings:

	Bucket:
		S3 bucket that data is written to
	Prefix (optional):
		prefix prepended to the S3 object name
		defaults to no prefix
	PrefixKey (optional):
		JSON key-value that is used as the prefix prepended to the S3 object name, overrides Prefix
		defaults to no prefix

When loaded with a factory, the sink uses this JSON configuration:

	{
		"type": "s3",
		"settings": {
			"bucket": "foo-bucket"
		}
	}
*/
type S3 struct {
	Bucket    string `json:"bucket"`
	Prefix    string `json:"prefix"`
	PrefixKey string `json:"prefix_key"`
}

// Send sinks a channel of encapsulated data with the S3 sink.
func (sink *S3) Send(ctx context.Context, ch *config.Channel) error {
	if !s3uploader.IsEnabled() {
		s3uploader.Setup()
	}

	files := make(map[string]*os.File)

	var prefix string
	if sink.Prefix != "" {
		prefix = sink.Prefix
	}

	// newline character for Unix-based systems
	separator := []byte("\n")

	for capsule := range ch.C {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if sink.PrefixKey != "" {
				prefix = capsule.Get(sink.PrefixKey).String()
			}

			if _, ok := files[prefix]; !ok {
				f, err := os.CreateTemp("", "substation")
				if err != nil {
					return fmt.Errorf("sink s3 bucket %s prefix %s: %v", sink.Bucket, prefix, err)
				}

				defer os.Remove(f.Name()) //nolint:staticcheck // SA9001: channel is closed on error, defer will run
				defer f.Close()           //nolint:staticcheck // SA9001: channel is closed on error, defer will run

				files[prefix] = f
			}

			if _, err := files[prefix].Write(capsule.Data()); err != nil {
				return fmt.Errorf("sink s3 bucket %s prefix %s: %v", sink.Bucket, prefix, err)
			}
			if _, err := files[prefix].Write(separator); err != nil {
				return fmt.Errorf("sink s3 bucket %s prefix %s: %v", sink.Bucket, prefix, err)
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
			return fmt.Errorf("sink s3 bucket %s: %v", sink.Bucket, err)
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
		if _, err := s3uploader.Upload(ctx, sink.Bucket, key, reader); err != nil {
			return fmt.Errorf("sink s3 bucket %s key %s: %v", sink.Bucket, key, err)
		}

		// s3uploader.Upload does not return the size of uploaded data, so we use the size of the uncompressed file when reporting stats for debugging
		fs, err := file.Stat()
		if err != nil {
			return fmt.Errorf("sink s3 bucket %s key %s: %v", sink.Bucket, key, err)
		}

		log.WithField(
			"bucket", sink.Bucket,
		).WithField(
			"key", key,
		).WithField(
			"size", fs.Size(),
		).Debug("uploaded data to S3")
	}

	return nil
}

/*
 createKey creates a date-based S3 object key that has this naming convention:
	[prefix : optional]/[year]/[month]/[day]/[uuid].gz
*/
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
