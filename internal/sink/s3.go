package sink

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/brexhq/substation/internal/aws/s3manager"
	"github.com/brexhq/substation/internal/log"
)

const layout = "2006-01-02"

/*
S3 sinks data as gzip compressed objects to an AWS S3 bucket. Object names contain the year, month, and day the data was processed by the sink; they can be optionally prefixed with a custom string.

The sink has these settings:
	Bucket:
		S3 bucket that data is written to
	Prefix (optional):
		prefix prepended to the S3 object name
		defaults to no prefix

The sink uses this Jsonnet configuration:
	{
		type: 's3',
		settings: {
			bucket: 'foo-bucket',
		},
	}
*/
type S3 struct {
	Bucket string `json:"bucket"`
	Prefix string `json:"prefix"`
}

var s3managerAPI s3manager.UploaderAPI

// Send sinks a channel of bytes with the S3 sink.
func (sink *S3) Send(ctx context.Context, ch chan []byte, kill chan struct{}) error {
	if !s3managerAPI.IsEnabled() {
		s3managerAPI.Setup()
	}

	// tracks inidividual data pulled from ch and written to the S3 object
	var count int

	// flushes the channel writing all data to a gzip buffer
	var buffer bytes.Buffer
	writer := gzip.NewWriter(&buffer)

	sep := []byte("\n")
	for data := range ch {
		select {
		case <-kill:
			return nil
		default:
			data = append(data, sep...)
			writer.Write(data)
			count++
		}
	}
	writer.Close()

	key := formatKey(sink.Prefix) + ".gz"
	if _, err := s3managerAPI.Upload(ctx, buffer.Bytes(), sink.Bucket, key); err != nil {
		return fmt.Errorf("err failed to upload data to bucket %s and key %s: %v", sink.Bucket, key, err)
	}

	log.WithField(
		"count", count,
	).WithField(
		"bucket", sink.Bucket,
	).WithField(
		"key", key,
	).Debug("uploaded data to S3")

	return nil
}

// formatPrefix creates an object key prefix based on the current time:
//  [prefix:optional]/[year]/[month]/[day]/[uuid]
func formatKey(prefix string) string {
	now := time.Now().Format(layout)
	var key string

	if prefix != "" {
		key = prefix + "/"
	}

	for _, s := range strings.Split(now, "-") {
		key += s + "/"
	}

	key += uuid.NewString()
	return key
}
