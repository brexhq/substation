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
S3 implements the Sink interface and writes gzip compressed objects to an S3 bucket. More information is available in the README.

Bucket: the S3 bucket that objects are written to
Prefix (optional): prefix prepended to the S3 object name
*/
type S3 struct {
	api    s3manager.UploaderAPI
	Bucket string `mapstructure:"bucket"`
	Prefix string `mapstructure:"prefix"`
}

// Send sends a channel of bytes to the S3 bucket defined by this sink.
func (sink *S3) Send(ctx context.Context, ch chan []byte, kill chan struct{}) error {
	if !sink.api.IsEnabled() {
		sink.api.Setup()
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
	if _, err := sink.api.Upload(ctx, buffer.Bytes(), sink.Bucket, key); err != nil {
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
