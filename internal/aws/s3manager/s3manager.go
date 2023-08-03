// package s3manager provides methods and functions for downloading and uploading objects in AWS S3.
package s3manager

import (
	"fmt"
	"io"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"
	"github.com/aws/aws-xray-sdk-go/xray"
	_aws "github.com/brexhq/substation/internal/aws"
	"github.com/brexhq/substation/internal/media"
)

// NewS3 returns a configured S3 client.
func NewS3(cfg _aws.Config) *s3.S3 {
	conf, sess := _aws.New(cfg)

	c := s3.New(sess, conf)
	if _, ok := os.LookupEnv("AWS_XRAY_DAEMON_ADDRESS"); ok {
		xray.AWS(c.Client)
	}

	return c
}

// NewS3Downloader returns a configured Downloader client.
func NewS3Downloader(cfg _aws.Config) *s3manager.Downloader {
	return s3manager.NewDownloaderWithClient(NewS3(cfg))
}

// DownloaderAPI wraps the Downloader API interface.
type DownloaderAPI struct {
	Client s3manageriface.DownloaderAPI
}

// Setup creates a new Downloader client.
func (a *DownloaderAPI) Setup(cfg _aws.Config) {
	a.Client = NewS3Downloader(cfg)
}

// IsEnabled returns true if the client is enabled and ready for use.
func (a *DownloaderAPI) IsEnabled() bool {
	return a.Client != nil
}

// Download is a convenience wrapper for downloading an object from S3.
func (a *DownloaderAPI) Download(ctx aws.Context, bucket, key string, dst io.WriterAt) (int64, error) {
	// TODO(v1.0.0): add ARN support
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	size, err := a.Client.DownloadWithContext(ctx, dst, input)
	if err != nil {
		return 0, fmt.Errorf("s3manager download bucket %s key %s: %v", bucket, key, err)
	}

	return size, nil
}

// NewS3Uploader returns a configured Uploader client.
func NewS3Uploader(cfg _aws.Config) *s3manager.Uploader {
	return s3manager.NewUploaderWithClient(NewS3(cfg))
}

// UploaderAPI wraps the Uploader API interface.
type UploaderAPI struct {
	Client s3manageriface.UploaderAPI
}

// Setup creates a new Uploader client.
func (a *UploaderAPI) Setup(cfg _aws.Config) {
	a.Client = NewS3Uploader(cfg)
}

// IsEnabled returns true if the client is enabled and ready for use.
func (a *UploaderAPI) IsEnabled() bool {
	return a.Client != nil
}

// Upload is a convenience wrapper for uploading an object to S3.
func (a *UploaderAPI) Upload(ctx aws.Context, bucket, key string, src io.Reader) (*s3manager.UploadOutput, error) {
	// temporary file is used so that the src can have its content identified and be uploaded to S3
	dst, err := os.CreateTemp("", "substation")
	if err != nil {
		return nil, fmt.Errorf("s3manager upload bucket %s key %s: %v", bucket, key, err)
	}
	defer os.Remove(dst.Name())
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return nil, fmt.Errorf("s3manager upload bucket %s key %s: %v", bucket, key, err)
	}

	mediaType, err := media.File(dst)
	if err != nil {
		return nil, fmt.Errorf("s3manager upload bucket %s key %s: %v", bucket, key, err)
	}

	if _, err := dst.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("s3manager upload bucket %s key %s: %v", bucket, key, err)
	}
	// TODO(v1.0.0): add ARN support
	input := &s3manager.UploadInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        dst,
		ContentType: aws.String(mediaType),
	}

	resp, err := a.Client.UploadWithContext(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("s3manager upload bucket %s key %s: %v", bucket, key, err)
	}

	return resp, nil
}
