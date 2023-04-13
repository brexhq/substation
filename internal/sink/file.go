package sink

import (
	"context"
	"fmt"
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
	files := make(map[string]*fw)

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
		// - extension (always .gz)
		path = cwd + "/" + time.Now().Format("2006/01/02") + "/" + uuid.New().String() + ".gz"
	}

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
				if err := os.MkdirAll(filepath.Dir(innerPath), 0o770); err != nil {
					return fmt.Errorf("sink: file: file_path %s: %v", path, err)
				}

				f, err := os.Create(innerPath)
				if err != nil {
					return fmt.Errorf("sink: file: file_path %s: %v", innerPath, err)
				}

				// TODO: make FileFormat configurable
				files[innerPath] = NewFileWrapper(f, config.Config{Type: "text_gzip"})
				defer files[innerPath].Close() //nolint:staticcheck // SA9001: channel is closed on error, defer will run
			}

			if _, err := files[innerPath].Write(capsule.Data()); err != nil {
				return fmt.Errorf("sink: file: file_path %s: %v", innerPath, err)
			}
		}
	}

	for path, file := range files {
		fs, err := file.Stat()
		if err != nil {
			return fmt.Errorf("sink: file: %v", err)
		}

		log.WithField(
			"path", path,
		).WithField(
			"size", fs.Size(),
		).WithField(
			"type", file.Type(),
		).Debug("wrote data to file")
	}

	return nil
}
