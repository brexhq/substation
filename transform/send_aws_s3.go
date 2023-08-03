package transform

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/aws"
	"github.com/brexhq/substation/internal/aws/s3manager"
	"github.com/brexhq/substation/internal/errors"
	mess "github.com/brexhq/substation/message"
)

type sendAWSS3Config struct {
	Auth    config.ConfigAWSAuth `json:"auth"`
	Request config.ConfigRequest `json:"request"`
	// Bucket is the AWS S3 bucket that data is written to.
	Bucket string `json:"bucket"`
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

type sendAWSS3 struct {
	conf sendAWSS3Config

	path      string
	extension string
	// client is safe for concurrent use.
	client s3manager.UploaderAPI
	// buffer is safe for concurrent use.
	mu     *sync.Mutex
	buffer map[string]*fw
}

func newSendAWSS3(_ context.Context, cfg config.Config) (*sendAWSS3, error) {
	conf := sendAWSS3Config{}
	if err := config.Decode(cfg.Settings, &conf); err != nil {
		return nil, err
	}

	// Validate required options.
	if conf.Bucket == "" {
		return nil, fmt.Errorf("transform: send_aws_s3: bucket stream: %v", errors.ErrMissingRequiredOption)
	}

	if conf.FileFormat.Type == "" {
		conf.FileFormat.Type = "json"
	}

	if conf.FileCompression.Type == "" {
		conf.FileCompression.Type = "gzip"
	}

	send := sendAWSS3{
		conf: conf,
	}

	// File extensions are dynamic and not directly configurable.
	send.extension = NewFileExtension(conf.FileFormat, conf.FileCompression)
	now := time.Now()

	// Default object key is: year/month/day/uuid.extension.
	send.path = conf.FilePath.New()
	if send.path == "" {
		send.path = path.Join(
			now.Format("2006"), now.Format("01"), now.Format("02"),
			uuid.New().String(),
		) + send.extension
	} else if conf.FilePath.Extension {
		send.path += send.extension
	}

	// Setup the AWS client.
	send.client.Setup(aws.Config{
		Region:     conf.Auth.Region,
		AssumeRole: conf.Auth.AssumeRole,
		MaxRetries: conf.Request.MaxRetries,
	})

	send.mu = &sync.Mutex{}
	send.buffer = make(map[string]*fw)

	return &send, nil
}

func (t *sendAWSS3) Close(context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	for _, f := range t.buffer {
		f.Close()
		os.Remove(f.Name())
	}

	return nil
}

func (t *sendAWSS3) Transform(ctx context.Context, messages ...*mess.Message) ([]*mess.Message, error) {
	// Lock the transform to prevent concurrent access to the buffer.
	t.mu.Lock()
	defer t.mu.Unlock()

	control := false
	for _, message := range messages {
		if message.IsControl() {
			control = true
			continue
		}

		// path is used so that key values can be interpolated into the path.
		path := t.path

		// If either prefix or suffix keys are set, then the object key is non-default
		// and can be safely interpolated. If either are empty strings, then an error
		// is returned.
		if t.conf.FilePath.PrefixKey != "" {
			prefix := message.Get(t.conf.FilePath.PrefixKey).String()
			if prefix == "" {
				return nil, fmt.Errorf("transform: send_aws_s3: bucket %s object %s: %v", t.conf.Bucket, path, fmt.Errorf("empty prefix string"))
			}

			path = strings.Replace(path, "${PATH_PREFIX}", prefix, 1)
		}
		if t.conf.FilePath.SuffixKey != "" {
			suffix := message.Get(t.conf.FilePath.SuffixKey).String()
			if suffix == "" {
				return nil, fmt.Errorf("transform: send_aws_s3: bucket %s object %s: %v", t.conf.Bucket, path, fmt.Errorf("empty suffix string"))
			}

			path = strings.Replace(path, "${PATH_SUFFIX}", suffix, 1)
		}

		if _, ok := t.buffer[path]; !ok {
			f, err := os.CreateTemp("", "substation")
			if err != nil {
				return nil, fmt.Errorf("transform: send_aws_s3: bucket %s object %s: %v", t.conf.Bucket, path, err)
			}

			if t.buffer[path], err = NewFileWrapper(f, t.conf.FileFormat, t.conf.FileCompression); err != nil {
				return nil, fmt.Errorf("send: file: file_path %s: %v", path, err)
			}
		}

		if _, err := t.buffer[path].Write(message.Data()); err != nil {
			return nil, fmt.Errorf("transform: send_aws_s3: bucket %s object %s: %v", t.conf.Bucket, path, err)
		}
	}

	// If a control message is received, then files are closed, uploaded to S3, and
	// removed from the buffer.
	if !control {
		return messages, nil
	}

	for path, file := range t.buffer {
		defer os.Remove(file.Name())

		// Flushes the file before uploading to S3.
		if err := file.Close(); err != nil {
			return nil, fmt.Errorf("transform: send_aws_s3: bucket %s object %s: %v", t.conf.Bucket, path, err)
		}

		// s3uploader requires an open file.
		f, err := os.Open(file.Name())
		if err != nil {
			return nil, fmt.Errorf("transform: send_aws_s3: bucket %s object %s: %v", t.conf.Bucket, path, err)
		}
		defer f.Close()

		if _, err := t.client.Upload(ctx, t.conf.Bucket, path, f); err != nil {
			return nil, fmt.Errorf("transform: send_aws_s3: bucket %s object %s: %v", t.conf.Bucket, path, err)
		}

		delete(t.buffer, path)
	}

	return messages, nil
}
