package transform

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
	AWS    iconfig.AWS      `json:"aws"`
	Retry  iconfig.Retry    `json:"retry"`

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

func (c *sendAWSS3Config) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *sendAWSS3Config) Validate() error {
	if c.Bucket == "" {
		return fmt.Errorf("bucket: %v", errors.ErrMissingRequiredOption)
	}

	if c.FileFormat.Type == "" {
		c.FileFormat.Type = "json"
	}

	if c.FileCompression.Type == "" {
		c.FileCompression.Type = "gzip"
	}

	return nil
}

func newSendAWSS3(_ context.Context, cfg config.Config) (*sendAWSS3, error) {
	conf := sendAWSS3Config{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: new_send_aws_s3: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: new_send_aws_s3: %v", err)
	}

	tf := sendAWSS3{
		conf: conf,
	}

	// File extensions are dynamic and not directly configurable.
	tf.extension = file.NewExtension(conf.FileFormat, conf.FileCompression)

	buffer, err := aggregate.New(conf.Buffer)
	if err != nil {
		return nil, fmt.Errorf("transform: new_send_aws_s3: %v", err)
	}
	tf.buffer = buffer

	// Setup the AWS client.
	tf.client.Setup(aws.Config{
		Region:     conf.AWS.Region,
		AssumeRole: conf.AWS.AssumeRole,
		MaxRetries: conf.Retry.Attempts,
	})

	return &tf, nil
}

type sendAWSS3 struct {
	conf sendAWSS3Config

	extension string
	// client is safe for concurrent use.
	client s3manager.UploaderAPI
	// buffer is safe for concurrent use.
	buffer *aggregate.Aggregate
}

func (tf *sendAWSS3) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		for key := range tf.buffer.GetAll() {
			if err := tf.writeFile(ctx, key); err != nil {
				return nil, fmt.Errorf("transform: send_aws_s3: %v", err)
			}
		}

		tf.buffer.ResetAll()
		return []*message.Message{msg}, nil
	}

	key := msg.GetValue(tf.conf.FilePath.PrefixKey).String()
	// Writes data as an object to S3 only when the buffer is full.
	if ok := tf.buffer.Add(key, msg.Data()); ok {
		return []*message.Message{msg}, nil
	}

	if err := tf.writeFile(ctx, key); err != nil {
		return nil, fmt.Errorf("transform: send_aws_s3: %v", err)
	}

	// Reset the buffer and add the msg data.
	tf.buffer.Reset(key)
	_ = tf.buffer.Add(key, msg.Data())

	return []*message.Message{msg}, nil
}

func (t *sendAWSS3) writeFile(ctx context.Context, prefix string) error {
	// If the buffer is empty, then there is nothing to write.
	if t.buffer.Count(prefix) == 0 {
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

	for _, data := range t.buffer.Get(prefix) {
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

func (tf *sendAWSS3) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}

func (*sendAWSS3) Close(context.Context) error {
	return nil
}
