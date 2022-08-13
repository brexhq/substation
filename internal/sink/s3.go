package sink

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jshlbrd/go-aggregate"

	"github.com/brexhq/substation/config"
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
	PrefixKey (optional):
		JSON key-value that is used as the prefix prepended to the S3 object name, overrides Prefix
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
	Bucket    string `json:"bucket"`
	Prefix    string `json:"prefix"`
	PrefixKey string `json:"prefix_key"`
}

var s3managerAPI s3manager.UploaderAPI

// Send sinks a channel of encapsulated data with the S3 sink.
func (sink *S3) Send(ctx context.Context, ch chan config.Capsule, kill chan struct{}) error {
	if !s3managerAPI.IsEnabled() {
		s3managerAPI.Setup()
	}

	buffer := map[string]*aggregate.Bytes{}

	var prefix string
	if sink.Prefix != "" {
		prefix = sink.Prefix
	}

	sep := []byte("\n")
	for cap := range ch {
		select {
		case <-kill:
			return nil
		default:
			if sink.PrefixKey != "" {
				prefix = cap.Get(sink.PrefixKey).String()
			}

			if _, ok := buffer[prefix]; !ok {
				// aggregate up to 10MB or 100,000 items
				buffer[prefix] = &aggregate.Bytes{}
				buffer[prefix].New(1000*1000*10, 100000)
			}

			// add data to the buffer
			// if buffer is full, then send the aggregated data
			ok, err := buffer[prefix].Add(cap.GetData())
			if err != nil {
				return fmt.Errorf("sink s3 bucket %s prefix %s: %v", sink.Bucket, prefix, err)
			}

			if !ok {
				var buf bytes.Buffer
				writer := gzip.NewWriter(&buf)
				items := buffer[prefix].Get()
				for _, i := range items {
					writer.Write(i)
					writer.Write(sep)
				}
				writer.Close()

				key := formatKey(prefix) + ".gz"
				if _, err := s3managerAPI.Upload(ctx, buf.Bytes(), sink.Bucket, key); err != nil {
					return fmt.Errorf("sink s3 bucket %s key %s: %v", sink.Bucket, key, err)
				}

				log.WithField(
					"count", buffer[prefix].Count(),
				).WithField(
					"bucket", sink.Bucket,
				).WithField(
					"key", key,
				).Debug("uploaded data to S3")

				buffer[prefix].Reset()
				buffer[prefix].Add(cap.GetData())
			}
		}
	}

	// iterate and send remaining buffers
	for prefix := range buffer {
		count := buffer[prefix].Count()
		if count == 0 {
			continue
		}

		var buf bytes.Buffer
		writer := gzip.NewWriter(&buf)
		items := buffer[prefix].Get()
		for _, b := range items {
			writer.Write(b)
			writer.Write(sep)
		}
		writer.Close()

		key := formatKey(prefix) + ".gz"
		if _, err := s3managerAPI.Upload(ctx, buf.Bytes(), sink.Bucket, key); err != nil {
			// Upload err returns metadata
			return fmt.Errorf("sink s3: %v", err)
		}

		log.WithField(
			"count", count,
		).WithField(
			"bucket", sink.Bucket,
		).WithField(
			"key", key,
		).Debug("uploaded data to S3")
	}

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
