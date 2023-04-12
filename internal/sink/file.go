package sink

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/log"
)

const (
	// errFileEmptyPrefix is returned when the sink is configured with a prefix
	// key, but the key is not found in the object or the key is empty.
	errFileEmptyPrefix = errors.Error("empty prefix string")
	// errFileEmptySuffix is returned when the sink is configured with a suffix
	// key, but the key is not found in the object or the key is empty.
	errFileEmptySuffix = errors.Error("empty suffix string")
)

// file sinks data as gzip compressed files to local disk.
type sinkFile struct {
	// FilePath determines how the name of the uploaded object is constructed.
	// One of these formats is constructed depending on the configuration:
	//
	// - prefix/date_format/uuid.extension
	//
	// - prefix/date_format/uuid/suffix.extension
	FilePath filePath `json:"file_path"`
}

// Send sinks a channel of encapsulated data with the sink.
func (s *sinkFile) Send(ctx context.Context, ch *config.Channel) error {
	files := make(map[string]*os.File)

	path := s.FilePath.New()
	if path == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("sink: file: %v", err)
		}

		// default path is:
		// - current directory
		// - year, month, and day
		// - random UUID
		path = cwd + "/" + time.Now().Format("2006/01/02") + "/" + uuid.New().String()

		// currently only supports gzip compression
		path += ".gz"
	}

	// newline character for Unix-based systems, not compatible with Windows
	separator := []byte("\n")

	for capsule := range ch.C {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// innerPath is used so that key values can be interpolated into the file path
			innerPath := path

			// if either prefix or suffix keys are set, then the object name is non-default
			// and can be safely interpolated. if either are empty strings, then an error
			// is returned.
			if s.FilePath.PrefixKey != "" {
				prefix := capsule.Get(s.FilePath.PrefixKey).String()
				if prefix == "" {
					return fmt.Errorf("sink: file: %v", errFileEmptyPrefix)
				}

				innerPath = strings.Replace(innerPath, "${PATH_PREFIX}", prefix, 1)
			}
			if s.FilePath.SuffixKey != "" {
				suffix := capsule.Get(s.FilePath.SuffixKey).String()
				if suffix == "" {
					return fmt.Errorf("sink: file: %v", errFileEmptySuffix)
				}

				innerPath = strings.Replace(innerPath, "${PATH_SUFFIX}", suffix, 1)
			}

			if _, ok := files[innerPath]; !ok {
				f, err := os.CreateTemp("", "substation")
				if err != nil {
					return fmt.Errorf("sink: file: file_path %s: %v", innerPath, err)
				}

				defer os.Remove(f.Name()) //nolint:staticcheck // SA9001: channel is closed on error, defer will run
				defer f.Close()           //nolint:staticcheck // SA9001: channel is closed on error, defer will run
				files[innerPath] = f
			}

			if _, err := files[innerPath].Write(capsule.Data()); err != nil {
				return fmt.Errorf("sink: file: file_path %s: %v", innerPath, err)
			}
			if _, err := files[innerPath].Write(separator); err != nil {
				return fmt.Errorf("sink: file: file_path %s: %v", innerPath, err)
			}
		}
	}

	for path, file := range files {
		if _, err := file.Seek(0, 0); err != nil {
			return fmt.Errorf("sink: file: %v", err)
		}

		if err := os.MkdirAll(filepath.Dir(path), 0o770); err != nil {
			return fmt.Errorf("sink: file: file_path %s: %v", path, err)
		}

		f, err := os.Create(path)
		if err != nil {
			fmt.Println("could not open file")
			return fmt.Errorf("sink: file: file_path %s: %v", path, err)
		}
		defer f.Close()

		reader, writer := io.Pipe()
		defer reader.Close()

		// goroutine avoids deadlock
		go func() {
			// currently only supports gzip compression
			gz := gzip.NewWriter(f)
			defer writer.Close()
			defer gz.Close()

			_, _ = io.Copy(gz, file)
		}()

		if _, err := io.Copy(f, reader); err != nil {
			return fmt.Errorf("sink: file: file_path %s: %v", path, err)
		}

		fs, err := f.Stat()
		if err != nil {
			return fmt.Errorf("sink: file: %v", err)
		}

		log.WithField(
			"file_path", path,
		).WithField(
			"size", fs.Size(),
		).Debug("wrote data to file")
	}

	return nil
}
