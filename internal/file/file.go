// package file provides functions that can be used to retrieve files from local and remote locations.
package file

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"

	iconfig "github.com/brexhq/substation/v2/internal/config"
	"github.com/brexhq/substation/v2/internal/http"
)

var (
	httpClient   http.HTTP
	s3downloader *manager.Downloader
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

			c := s3.NewFromConfig(awsCfg)
			s3downloader = manager.NewDownloader(c)
		}

		// "s3://bucket/key" becomes ["bucket" "key"]
		paths := strings.SplitN(strings.TrimPrefix(location, "s3://"), "/", 2)

		// Download the file from S3.
		ctx = context.WithoutCancel(ctx)
		size, err := s3downloader.Download(ctx, dst, &s3.GetObjectInput{
			Bucket: &paths[0],
			Key:    &paths[1],
		})
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
