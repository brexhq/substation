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
	_config "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/file"
	mess "github.com/brexhq/substation/message"
)

type sendAWSS3Config struct {
	Buffer  aggregate.Config      `json:"buffer"`
	Auth    _config.ConfigAWSAuth `json:"auth"`
	Request _config.ConfigRequest `json:"request"`
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
	if err := _config.Decode(cfg.Settings, &conf); err != nil {
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
	send.extension = file.NewExtension(conf.FileFormat, conf.FileCompression)
	send.mu = sync.Mutex{}
	send.buffer = make(map[string]*aggregate.Aggregate)
	send.bufferCfg = aggregate.Config{
		Count:    conf.Buffer.Count,
		Size:     conf.Buffer.Size,
		Interval: conf.Buffer.Interval,
	}

	// Setup the AWS client.
	send.client.Setup(aws.Config{
		Region:     conf.Auth.Region,
		AssumeRole: conf.Auth.AssumeRole,
		MaxRetries: conf.Request.MaxRetries,
	})

	return &send, nil
}

func (*sendAWSS3) Close(context.Context) error {
	return nil
}

func (send *sendAWSS3) Transform(ctx context.Context, message *mess.Message) ([]*mess.Message, error) {
	// Lock the transform to prevent concurrent access to the buffer.
	send.mu.Lock()
	defer send.mu.Unlock()

	if message.IsControl() {
		for prefixKey := range send.buffer {
			if err := send.writeFile(ctx, prefixKey); err != nil {
				return nil, fmt.Errorf("transform: send_file: %v", err)
			}
		}

		send.buffer = make(map[string]*aggregate.Aggregate)
		return []*mess.Message{message}, nil
	}

	var prefixKey string
	if send.conf.FilePath.PrefixKey != "" {
		prefixKey = message.Get(send.conf.FilePath.PrefixKey).String()
	}

	if _, ok := send.buffer[prefixKey]; !ok {
		agg, err := aggregate.New(send.bufferCfg)
		if err != nil {
			return nil, fmt.Errorf("transform: send_file: %v", err)
		}

		send.buffer[prefixKey] = agg
	}

	// Writes data as an object to S3 only when the buffer is full.
	if ok := send.buffer[prefixKey].Add(message.Data()); ok {
		return []*mess.Message{message}, nil
	}

	if err := send.writeFile(ctx, prefixKey); err != nil {
		return nil, fmt.Errorf("transform: send_file: %v", err)
	}

	// Reset the buffer and add the message data.
	send.buffer[prefixKey].Reset()
	_ = send.buffer[prefixKey].Add(message.Data())

	return []*mess.Message{message}, nil
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
