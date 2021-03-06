package s3manager

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"
	"github.com/aws/aws-xray-sdk-go/xray"
)

const scannerMaxCapacity = 1024 * 1024 * 100

//NewS3 creates a new session connection to S3
func NewS3() *s3.S3 {
	conf := aws.NewConfig()

	// provides forward compatibility for the Go SDK to support env var configuration settings
	// https://github.com/aws/aws-sdk-go/issues/4207
	max, found := os.LookupEnv("AWS_MAX_ATTEMPTS")
	if found {
		m, err := strconv.Atoi(max)
		if err != nil {
			panic(err)
		}

		conf = conf.WithMaxRetries(m)
	}

	c := s3.New(
		session.Must(session.NewSession()),
		conf,
	)

	if _, ok := os.LookupEnv("AWS_XRAY_DAEMON_ADDRESS"); ok {
		xray.AWS(c.Client)
	}

	return c
}

// DownloaderAPI wraps a S3 client interface
type DownloaderAPI struct {
	Client s3manageriface.DownloaderAPI
}

// NewS3Downloader returns an XRay-enabled S3 Downloader client
func NewS3Downloader() *s3manager.Downloader {
	return s3manager.NewDownloaderWithClient(NewS3())
}

// Setup initialiazes Client with an XRay-enabled S3Uploader client
func (a *DownloaderAPI) Setup() {
	a.Client = NewS3Downloader()
}

//IsEnabled checks if the client has been set
func (a *DownloaderAPI) IsEnabled() bool {
	return a.Client != nil
}

// Download wraps AWS S3Downloader DownloadWithContext API with optional XRay support
func (a *DownloaderAPI) Download(ctx aws.Context, bucket, key string) ([]byte, int64, error) {
	buf := aws.NewWriteAtBuffer([]byte{})
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	size, err := a.Client.DownloadWithContext(ctx, buf, input)
	if err != nil {
		return nil, 0, fmt.Errorf("download bucket %s key %s: %w", bucket, key, err)
	}
	return buf.Bytes(), size, nil
}

// DownloadAsScanner wraps DownloaderClient.DownloadWithContext and returns object as a decoded scanner
func (a *DownloaderAPI) DownloadAsScanner(ctx aws.Context, bucket, key string) (*bufio.Scanner, error) {
	buf, size, err := a.Download(ctx, bucket, key)
	if err != nil {
		return nil, err
	}

	if size == 0 {
		return nil, nil
	}

	ct := http.DetectContentType(buf)
	decoded, err := decode(buf, ct)
	if err != nil {
		return nil, fmt.Errorf("decode bucket %s key %s: %v", bucket, key, err)
	}

	s := createScanner(decoded)
	return s, nil
}

// decode converts bytes into a decoded io.Reader
func decode(buf []byte, contentType string) (io.Reader, error) {
	switch t := contentType; t {
	case "application/x-gzip":
		content, err := gzip.NewReader(bytes.NewBuffer(buf))
		if err != nil {
			return nil, fmt.Errorf("decode content type %s: %v", contentType, err)
		}
		return content, nil
	default:
		return bytes.NewBuffer(buf), nil
	}
}

// createScanner creates a bufio.Scanner from an io.Reader
func createScanner(content io.Reader) *bufio.Scanner {
	scanner := bufio.NewScanner(content)
	b := make([]byte, scannerMaxCapacity)
	scanner.Buffer(b, scannerMaxCapacity)
	return scanner
}

//NewS3Uploader creates a new S3 uploader client connection
func NewS3Uploader() *s3manager.Uploader {
	return s3manager.NewUploaderWithClient(NewS3())
}

// UploaderAPI wraps a S3 client interface
type UploaderAPI struct {
	Client s3manageriface.UploaderAPI
}

// IsEnabled checks if the client has been set
func (a *UploaderAPI) IsEnabled() bool {
	return a.Client != nil
}

// Setup initialiazes Client with an XRay-enabled S3Uploader client
func (a *UploaderAPI) Setup() {
	a.Client = NewS3Uploader()
}

// Upload wraps AWS S3manager's UploadInput API
func (a *UploaderAPI) Upload(ctx aws.Context, buffer []byte, bucket, key string) (*s3manager.UploadOutput, error) {
	input := &s3manager.UploadInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(buffer),
		ContentType: aws.String(http.DetectContentType(buffer)),
	}

	resp, err := a.Client.UploadWithContext(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("upload bucket %s key %s: %w", bucket, key, err)
	}
	return resp, nil
}
