package transform

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/aggregate"
	"github.com/brexhq/substation/internal/aws"
	"github.com/brexhq/substation/internal/aws/s3manager"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/file"
	"github.com/brexhq/substation/message"
)

type sendAWSS3Config struct {
	Buffer aggregate.Config `json:"buffer"`
	AWS    configAWS        `json:"aws"`
	Retry  configRetry      `json:"retry"`

	// Bucket is the AWS S3 bucket that data is written to.
	Bucket string `json:"bucket"`
	// FilePath determines how the name of the uploaded object is constructed.
	// See filePath.New for more information.
	FilePath file.Path `json:"file_path"`
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

	extension string
	// client is safe for concurrent use.
	client s3manager.UploaderAPI
	// buffer is safe for concurrent use.
	mu        sync.Mutex
	buffer    map[string]*aggregate.Aggregate
	bufferCfg aggregate.Config
}

func newSendAWSS3(_ context.Context, cfg config.Config) (*sendAWSS3, error) {
	conf := sendAWSS3Config{}
	if err := iconfig.Decode(cfg.Settings, &conf); err != nil {
		return nil, fmt.Errorf("transform: new_send_aws_s3: %v", err)
	}

	// Validate required options.
	if conf.Bucket == "" {
		return nil, fmt.Errorf("transform: new_send_aws_s3: bucket: %v", errors.ErrMissingRequiredOption)
	}

	if conf.FileFormat.Type == "" {
		conf.FileFormat.Type = "json"
	}

	if conf.FileCompression.Type == "" {
		conf.FileCompression.Type = "gzip"
	}

	tf := sendAWSS3{
		conf: conf,
	}

	// File extensions are dynamic and not directly configurable.
	tf.extension = file.NewExtension(conf.FileFormat, conf.FileCompression)
	tf.mu = sync.Mutex{}
	tf.buffer = make(map[string]*aggregate.Aggregate)
	tf.bufferCfg = aggregate.Config{
		Count:    conf.Buffer.Count,
		Size:     conf.Buffer.Size,
		Duration: conf.Buffer.Duration,
	}

	// Setup the AWS client.
	tf.client.Setup(aws.Config{
		Region:     conf.AWS.Region,
		AssumeRole: conf.AWS.AssumeRole,
		MaxRetries: conf.Retry.Attempts,
	})

	return &tf, nil
}

func (*sendAWSS3) Close(context.Context) error {
	return nil
}

func (tf *sendAWSS3) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	// Lock the transform to prevent concurrent access to the buffer.
	tf.mu.Lock()
	defer tf.mu.Unlock()

	if msg.IsControl() {
		for prefixKey := range tf.buffer {
			if err := tf.writeFile(ctx, prefixKey); err != nil {
				return nil, fmt.Errorf("transform: send_aws_s3: %v", err)
			}
		}

		tf.buffer = make(map[string]*aggregate.Aggregate)
		return []*message.Message{msg}, nil
	}

	var prefixKey string
	if tf.conf.FilePath.PrefixKey != "" {
		prefixKey = msg.GetObject(tf.conf.FilePath.PrefixKey).String()
	}

	if _, ok := tf.buffer[prefixKey]; !ok {
		agg, err := aggregate.New(tf.bufferCfg)
		if err != nil {
			return nil, fmt.Errorf("transform: send_aws_s3: %v", err)
		}

		tf.buffer[prefixKey] = agg
	}

	// Writes data as an object to S3 only when the buffer is full.
	if ok := tf.buffer[prefixKey].Add(msg.Data()); ok {
		return []*message.Message{msg}, nil
	}

	if err := tf.writeFile(ctx, prefixKey); err != nil {
		return nil, fmt.Errorf("transform: send_aws_s3: %v", err)
	}

	// Reset the buffer and add the msg data.
	tf.buffer[prefixKey].Reset()
	_ = tf.buffer[prefixKey].Add(msg.Data())

	return []*message.Message{msg}, nil
}

func (t *sendAWSS3) writeFile(ctx context.Context, prefix string) error {
	// If the buffer is empty, then there is nothing to write.
	if t.buffer[prefix].Count() == 0 {
		return nil
	}

	fpath := t.conf.FilePath.New()
	if fpath == "" {
		return fmt.Errorf("file_path is empty")
	}

	if prefix != "" {
		fpath = strings.Replace(fpath, "${PATH_PREFIX}", prefix, 1)
	}

	fpath += t.extension

	// Ensures that the path is OS agnostic.
	fpath = filepath.FromSlash(fpath)

	temp, err := os.CreateTemp("", "substation")
	if err != nil {
		return err
	}
	defer temp.Close()

	w, err := file.NewWrapper(temp, t.conf.FileFormat, t.conf.FileCompression)
	if err != nil {
		return err
	}

	for _, data := range t.buffer[prefix].Get() {
		if _, err := w.Write(data); err != nil {
			return err
		}
	}

	// Flush the file before uploading to S3.
	if err := w.Close(); err != nil {
		return err
	}

	f, err := os.Open(temp.Name())
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := t.client.Upload(ctx, t.conf.Bucket, fpath, f); err != nil {
		return err
	}

	return nil
}
