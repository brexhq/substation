// package file provides functions that can be used to retrieve files from local and remote locations.
package file

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/aws"
	"github.com/brexhq/substation/internal/aws/s3manager"
	"github.com/brexhq/substation/internal/http"
	"github.com/google/uuid"
	"github.com/klauspost/compress/snappy"
	"github.com/klauspost/compress/zstd"
)

var (
	httpClient   http.HTTP
	s3downloader s3manager.DownloaderAPI
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

	if strings.HasPrefix(location, "s3://") {
		if !s3downloader.IsEnabled() {
			s3downloader.Setup(aws.Config{})
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

type Wrapper struct {
	*os.File
	w       io.WriteCloser
	newline []byte
}

func (w *Wrapper) Write(b []byte) (int, error) {
	if w.newline != nil {
		b = append(b, w.newline...)
	}

	if w.w != nil {
		return w.w.Write(b)
	}

	return w.File.Write(b)
}

func (w *Wrapper) Close() error {
	if w.w != nil {
		if err := w.w.Close(); err != nil {
			return err
		}
	}

	return w.File.Close()
}

func NewWrapper(f *os.File, format config.Config, compression config.Config) (*Wrapper, error) {
	var newline []byte
	switch runtime.GOOS {
	case "windows":
		newline = []byte("\r\n")
	default:
		newline = []byte("\n")
	}

	// if the file format is not text-based, then newline is unused.
	// if a file format uses a specific compression, then it should
	// be configured and returned in this switch.
	switch format.Type {
	case "data":
		newline = nil
	}

	switch compression.Type {
	case "gzip":
		return &Wrapper{f, gzip.NewWriter(f), newline}, nil
	case "snappy":
		return &Wrapper{f, snappy.NewBufferedWriter(f), newline}, nil
	case "zstd":
		// TODO: Add settings support.
		z, err := zstd.NewWriter(f)
		if err != nil {
			return nil, err
		}

		return &Wrapper{f, z, newline}, nil
	default:
		return &Wrapper{f, nil, newline}, nil
	}
}

type Path struct {
	// Prefix is a prefix prepended to the file path.
	//
	// This is optional and has no default.
	Prefix string `json:"prefix"`
	// PrefixKey retrieves a value from an object that is used as
	// the prefix prepended to the file path. If used, then
	// this overrides Prefix.
	//
	// This is optional and has no default.
	PrefixKey string `json:"prefix_key"`
	// TimeFormat inserts a formatted datetime string into the file path.
	// Must be one of:
	//
	// - pattern-based layouts (https://gobyexample.com/procTime-formatting-parsing)
	//
	// - unix: epoch (supports fractions of a second)
	//
	// - unix_milli: epoch milliseconds
	//
	// This is optional and has no default.
	TimeFormat string `json:"time_format"`
	// UUID inserts a random UUID into the file path. If a suffix is
	// not set, then this is used as the filename.
	//
	// This is optional and defaults to false.
	UUID bool `json:"uuid"`
	// Extension appends a file extension to the filename.
	//
	// This is optional and defaults to false.
	Extension bool `json:"extension"`
}

// New constructs a file path using the pattern
// [prefix]/[prefix_key]/[time_format]/[uuid], where each field is optional
// and builds on the previous field. The caller is responsible for
// creating an OS agnostic file path (filepath.FromSlash is recommended).
//
// If only one field is set, then this constructs a filename,
// otherwise it constructs a file path.
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

	if p.PrefixKey != "" {
		arr = append(arr, "${PATH_PREFIX}")
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

	// if only one field is set, then this returns a filename, otherwise
	// it returns a file path.
	return path.Join(arr...)
}

// NewExtension returns a file extension based on file format and
// compression settings. The file extensions constructed by this function
// match this regular expression: `(\.json|\.txt)?(\.gz|\.zst)?`.
func NewExtension(format config.Config, compression config.Config) (ext string) {
	switch format.Type {
	case "data":
		break
	case "json":
		ext = ".json"
	case "text":
		ext = ".txt"
	}

	switch compression.Type {
	case "gzip":
		ext += ".gz"
	case "snappy":
		break
	case "zstd":
		ext += ".zst"
	}

	return ext
}
