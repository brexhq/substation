// package file provides functions that can be used to retrieve files from local and remote locations.
package file

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/brexhq/substation/internal/aws/s3manager"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/http"
)

var (
	httpClient   http.HTTP
	s3downloader s3manager.DownloaderAPI
)

// errEmptyFile is returned when Get is called but finds an empty file.
const errEmptyFile = errors.Error("empty file found")

// errNoFile is returned when Get is called but no file is found.
const errNoFile = errors.Error("no file found")

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

	if strings.HasPrefix(location, "s3://") {
		if !s3downloader.IsEnabled() {
			s3downloader.Setup()
		}

		// "s3://bucket/key" becomes ["bucket" "key"]
		paths := strings.SplitN(strings.TrimPrefix(location, "s3://"), "/", 2)

		size, err := s3downloader.Download(ctx, paths[0], paths[1], dst)
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
