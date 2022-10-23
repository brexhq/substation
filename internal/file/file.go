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
	s3managerAPI s3manager.DownloaderAPI
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

If a file is found, then it is saved to disk and the path is returned. The caller is responsible for removing files when they are no longer needed.
*/
func Get(ctx context.Context, path string) (string, error) {
	file, err := os.CreateTemp("", "tempfile")
	if err != nil {
		return "", fmt.Errorf("file %s: %v", path, err)
	}

	if _, err := os.Stat(path); err == nil {
		buf, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("file %s: %v", path, errEmptyFile)
		}

		if len(buf) == 0 {
			return "", fmt.Errorf("file %s: %v", path, errEmptyFile)
		}

		if _, err := file.Write(buf); err != nil {
			return "", fmt.Errorf("file %s: %v", path, err)
		}

		return file.Name(), nil
	}

	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") { //nolint:nestif // err checking
		if !httpClient.IsEnabled() {
			httpClient.Setup()
		}

		resp, err := httpClient.Get(ctx, path)
		if err != nil {
			return "", fmt.Errorf("file %s: %v", path, err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("file %s: %v", path, err)
		}

		if len(body) == 0 {
			return "", fmt.Errorf("file %s: %v", path, errEmptyFile)
		}

		if _, err := file.Write(body); err != nil {
			return "", fmt.Errorf("file %s: %v", path, err)
		}

		return file.Name(), nil
	}

	if strings.HasPrefix(path, "s3://") {
		if !s3managerAPI.IsEnabled() {
			s3managerAPI.Setup()
		}

		// "s3://bucket/key" becomes ["bucket" "key"]
		paths := strings.SplitN(strings.TrimPrefix(path, "s3://"), "/", 2)
		buf, size, err := s3managerAPI.Download(ctx, paths[0], paths[1])
		if err != nil {
			return "", fmt.Errorf("file %s: %v", path, err)
		}

		if size == 0 {
			return "", fmt.Errorf("file %s: %v", path, errEmptyFile)
		}

		if _, err := file.Write(buf); err != nil {
			return "", fmt.Errorf("file %s: %v", path, err)
		}

		return file.Name(), nil
	}

	return "", fmt.Errorf("file %s: %v", path, errNoFile)
}
