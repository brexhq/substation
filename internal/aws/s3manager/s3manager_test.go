package s3manager

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"
)

type mockedDownload struct {
	s3manageriface.DownloaderAPI
	Resp int64
}

func (m mockedDownload) DownloadWithContext(ctx aws.Context, w io.WriterAt, input *s3.GetObjectInput, options ...func(*s3manager.Downloader)) (int64, error) {
	return m.Resp, nil
}

func TestDownload(t *testing.T) {
	tests := []struct {
		resp  int64
		input struct {
			bucket string
			key    string
		}
		expected int64
	}{
		{
			resp: 1,
			input: struct {
				bucket string
				key    string
			}{
				bucket: "foo",
				key:    "bar",
			},
			expected: 1,
		},
	}

	ctx := context.TODO()

	for _, test := range tests {
		a := DownloaderAPI{
			mockedDownload{Resp: test.resp},
		}

		var dst io.WriterAt
		size, err := a.Download(ctx, test.input.bucket, test.input.key, dst)
		if err != nil {
			t.Fatalf("%d, unexpected error", err)
		}

		if size != test.expected {
			t.Errorf("expected %d, got %d", size, test.expected)
		}
	}
}

type mockedUpload struct {
	s3manageriface.UploaderAPI
	Resp s3manager.UploadOutput
}

func (m mockedUpload) UploadWithContext(ctx aws.Context, input *s3manager.UploadInput, options ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error) {
	return &m.Resp, nil
}

func TestUpload(t *testing.T) {
	tests := []struct {
		resp  s3manager.UploadOutput
		input struct {
			buffer []byte
			bucket string
			key    string
		}
		expected string
	}{
		{
			resp: s3manager.UploadOutput{
				Location: "foo",
			},
			input: struct {
				buffer []byte
				bucket string
				key    string
			}{
				buffer: []byte("foo"),
				bucket: "bar",
				key:    "baz",
			},
			expected: "foo",
		},
	}

	ctx := context.TODO()

	for _, test := range tests {
		a := UploaderAPI{
			mockedUpload{Resp: test.resp},
		}

		src := strings.NewReader("foo")
		resp, err := a.Upload(ctx, test.input.bucket, test.input.key, src)
		if err != nil {
			t.Fatalf("%d, unexpected error", err)
		}

		if resp.Location != test.expected {
			t.Errorf("expected %s, got %s", resp.Location, test.expected)
		}
	}
}
