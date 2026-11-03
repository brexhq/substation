// package file provides functions that can be used to retrieve files from local and remote locations.
package file

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
	"github.com/google/uuid"

	iconfig "github.com/brexhq/substation/v2/internal/config"
	"github.com/brexhq/substation/v2/internal/http"
)

var (
	httpClient    http.HTTP
	s3client      *s3.Client
	s3downloader  *manager.Downloader
	gcpdownloader *storage.Client
)

// errEmptyFile is returned when Get is called but finds an empty file.
var errEmptyFile = fmt.Errorf("empty file found")

// errNoFile is returned when Get is called but no file is found.
var errNoFile = fmt.Errorf("no file found")

/*
Get retrieves a file from these locations (in order):

- Local disk

- HTTP or HTTPS URL

- AWS S3

- GCP Storage

If a file is found, then it is saved as a temporary local file and the name is returned. The caller is responsible for removing files when they are no longer needed; files should be removed even if an error occurs.
*/
func Get(ctx context.Context, location string) (string, error) {
	dst, err := os.CreateTemp("", "substation")
	if err != nil {
		return "", fmt.Errorf("get %s: %v", location, err)
	}
	defer dst.Close()

	if _, err := os.Stat(location); err == nil {
		src, err := os.Open(location)
		if err != nil {
			return dst.Name(), fmt.Errorf("get %s: %v", location, err)
		}
		defer src.Close()

		size, err := io.Copy(dst, src)
		if err != nil {
			return dst.Name(), fmt.Errorf("get %s: %v", location, err)
		}

		if size == 0 {
			return dst.Name(), fmt.Errorf("get %s: %v", location, errEmptyFile)
		}

		return dst.Name(), nil
	}

	if strings.HasPrefix(location, "http://") || strings.HasPrefix(location, "https://") {
		if !httpClient.IsEnabled() {
			httpClient.Setup()
		}

		resp, err := httpClient.Get(ctx, location)
		if err != nil {
			return dst.Name(), fmt.Errorf("get %s: %v", location, err)
		}
		defer resp.Body.Close()

		size, err := io.Copy(dst, resp.Body)
		if err != nil {
			return dst.Name(), fmt.Errorf("get %s: %v", location, err)
		}

		if size == 0 {
			return dst.Name(), fmt.Errorf("get %s: %v", location, errEmptyFile)
		}

		return dst.Name(), nil
	}

	//nolint: nestif // ignore nesting complexity
	if strings.HasPrefix(location, "s3://") {
		if s3downloader == nil {
			awsCfg, err := iconfig.NewAWS(ctx, iconfig.AWS{})
			if err != nil {
				return dst.Name(), fmt.Errorf("get %s: %v", location, err)
			}

			s3client = s3.NewFromConfig(awsCfg)
			s3downloader = manager.NewDownloader(s3client)
		}

		// "s3://bucket/key" becomes ["bucket" "key"]
		paths := strings.SplitN(strings.TrimPrefix(location, "s3://"), "/", 2)

		// Download the file from S3.
		ctx = context.WithoutCancel(ctx)
		size, err := downloadS3(ctx, s3client, s3downloader, dst, paths[0], paths[1])
		if err != nil {
			return dst.Name(), fmt.Errorf("get %s: %v", location, err)
		}

		if size == 0 {
			return dst.Name(), fmt.Errorf("get %s: %v", location, errEmptyFile)
		}

		return dst.Name(), nil
	}

	//nolint: nestif // ignore nesting complexity
	if strings.HasPrefix(location, "gs://") {
		if gcpdownloader == nil {
			gcpdownloader, err = storage.NewClient(ctx)
			if err != nil {
				return dst.Name(), fmt.Errorf("get %s: %v", location, err)
			}
		}

		// "gs://bucket/key" becomes ["bucket" "key"]
		paths := strings.SplitN(strings.TrimPrefix(location, "gs://"), "/", 2)

		// Download the file from GCP Storage.
		ctx = context.WithoutCancel(ctx)
		reader, err := gcpdownloader.Bucket(paths[0]).Object(paths[1]).NewReader(ctx)
		if err != nil {
			return dst.Name(), fmt.Errorf("get %s: %v", location, err)
		}
		defer reader.Close()

		size, err := io.Copy(dst, reader)
		if err != nil {
			return dst.Name(), fmt.Errorf("get %s: %v", location, err)
		}

		if size == 0 {
			return dst.Name(), fmt.Errorf("get %s: %v", location, errEmptyFile)
		}

		return dst.Name(), nil
	}

	return dst.Name(), fmt.Errorf("get %s: %v", location, errNoFile)
}

type Path struct {
	// Prefix prepends a value to the file path.
	//
	// This is optional and has no default.
	Prefix string `json:"prefix"`
	// TimeFormat inserts a formatted datetime string into the file path.
	// Must be one of:
	//   - pattern-based layouts (https://gobyexample.com/procTime-formatting-parsing)
	//   - unix: epoch (supports fractions of a second)
	//   - unix_milli: epoch milliseconds
	//
	// This is optional and has no default.
	TimeFormat string `json:"time_format"`
	// UUID inserts a random UUID into the file path. If a suffix is
	// not set, then this is used as the filename.
	//
	// This is optional and defaults to false.
	UUID bool `json:"uuid"`
	// Suffix appends a value to the file path.
	//
	// This is optional and has no default.
	Suffix string `json:"suffix"`
}

// New constructs a file path using the pattern
// [p.Prefix]/[p.TimeFormat]/[p.UUID][p.Suffix], where each field is
// optional and builds on the previous field. The caller is responsible for
// creating an OS agnostic file path (filepath.FromSlash is recommended).
//
// If only one field is set, then this constructs a filename, otherwise it
// constructs a file path.
//
// If the struct is empty, then this returns an empty string. The caller is
// responsible for creating a default file path if needed.
func (p Path) New() string {
	// temporarily storing values for the file path in an array allows for any
	// individual field to be used as the filename if no other fields are set.
	arr := []string{}

	if p.Prefix != "" {
		arr = append(arr, p.Prefix)
	}

	if p.TimeFormat != "" {
		now := time.Now()

		// these options mirror process/time.go
		switch p.TimeFormat {
		case "unix":
			arr = append(arr, fmt.Sprintf("%d", now.Unix()))
		case "unix_milli":
			arr = append(arr, fmt.Sprintf("%d", now.UnixMilli()))
		default:
			arr = append(arr, now.Format(p.TimeFormat))
		}
	}

	if p.UUID {
		arr = append(arr, uuid.NewString())
	}

	return fmt.Sprintf("%s%s", path.Join(arr...), p.Suffix)
}

// s3DownloadAttempts is how many times downloadS3 restarts when the object
// changes mid-download.
const s3DownloadAttempts = 3

// s3Downloader downloads an S3 object in parallel ranged GETs.
type s3Downloader interface {
	Download(context.Context, io.WriterAt, *s3.GetObjectInput, ...func(*manager.Downloader)) (int64, error)
}

// s3HeadClient reads S3 object metadata.
type s3HeadClient interface {
	HeadObject(context.Context, *s3.HeadObjectInput, ...func(*s3.Options)) (*s3.HeadObjectOutput, error)
}

// downloadS3 downloads one version of an S3 object.
//
// The downloader fetches parts in parallel, each as its own ranged GET, so an
// object overwritten mid-download would otherwise be stitched together from
// two versions. Pinning every part to the ETag from HeadObject makes S3 reject
// a changed object with 412 Precondition Failed, and the download starts over.
func downloadS3(ctx context.Context, client s3HeadClient, d s3Downloader, dst *os.File, bucket, key string) (int64, error) {
	var err error
	for i := 0; i < s3DownloadAttempts; i++ {
		var head *s3.HeadObjectOutput
		head, err = client.HeadObject(ctx, &s3.HeadObjectInput{
			Bucket: &bucket,
			Key:    &key,
		})
		if err != nil {
			return 0, err
		}

		// Drop any bytes from an earlier attempt.
		if err = dst.Truncate(0); err != nil {
			return 0, err
		}

		var size int64
		size, err = d.Download(ctx, dst, &s3.GetObjectInput{
			Bucket:  &bucket,
			Key:     &key,
			IfMatch: head.ETag,
		})
		if err == nil {
			return size, nil
		}

		var apiErr smithy.APIError
		if !errors.As(err, &apiErr) || apiErr.ErrorCode() != "PreconditionFailed" {
			return 0, err
		}
	}

	return 0, fmt.Errorf("object changed during %d download attempts: %v", s3DownloadAttempts, err)
}
