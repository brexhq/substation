package sink

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/log"
)

// file sinks data as files to local disk.
type sinkFile struct {
	// FilePath determines how the name of the file is constructed.
	// See filePath.New for more information.
	FilePath filePath `json:"file_path"`
	// FileFormat determines the format of the file. These file formats are
	// supported:
	//
	// - json
	//
	// - text
	//
	// - data (binary data)
	//
	// If the format type does not have a common file extension, then
	// no extension is added to the file name.
	//
	// Defaults to json.
	FileFormat config.Config `json:"file_format"`
	// FileCompression determines the compression type applied to the file.
	// These compression types are supported:
	//
	// - gzip (https://en.wikipedia.org/wiki/Gzip)
	//
	// - snappy (https://en.wikipedia.org/wiki/Snappy_(compression))
	//
	// - zstd (https://en.wikipedia.org/wiki/Zstd)
	//
	// If the compression type does not have a common file extension, then
	// no extension is added to the file name.
	//
	// Defaults to gzip.
	FileCompression config.Config `json:"file_compression"`
}

// Send sinks a channel of encapsulated data with the sink.
func (s *sinkFile) Send(ctx context.Context, ch *config.Channel) error {
	files := make(map[string]*fw)

	// TODO: move to constructor
	if s.FileFormat.Type == "" {
		s.FileFormat.Type = "json"
	}

	// TODO: move to constructor
	if s.FileCompression.Type == "" {
		s.FileCompression.Type = "gzip"
	}

	// file extensions are dynamic and not directly configurable
	extension := NewFileExtension(s.FileFormat, s.FileCompression)
	now := time.Now()

	// default file path is: cwd/year/month/day/uuid.extension
	fpath := s.FilePath.New()
	if fpath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("sink: file: %v", err)
		}

		fpath = path.Join(
			cwd,
			now.Format("2006"), now.Format("01"), now.Format("02"),
			uuid.New().String(),
		) + extension
	} else if s.FilePath.Extension {
		fpath += extension
	}

	// ensures that the path is OS agnostic
	fpath = filepath.FromSlash(fpath)

	for capsule := range ch.C {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// innerPath is used so that key values can be interpolated into the file path
			innerPath := fpath

			// if either prefix or suffix keys are set, then the object name is non-default
			// and can be safely interpolated. if either are empty strings, then an error
			// is returned.
			if s.FilePath.PrefixKey != "" {
				prefix := capsule.Get(s.FilePath.PrefixKey).String()
				if prefix == "" {
					return fmt.Errorf("sink: file: %v", errEmptyPrefix)
				}

				innerPath = strings.Replace(innerPath, "${PATH_PREFIX}", prefix, 1)
			}
			if s.FilePath.SuffixKey != "" {
				suffix := capsule.Get(s.FilePath.SuffixKey).String()
				if suffix == "" {
					return fmt.Errorf("sink: file: %v", errEmptySuffix)
				}

				innerPath = strings.Replace(innerPath, "${PATH_SUFFIX}", suffix, 1)
			}

			if _, ok := files[innerPath]; !ok {
				if err := os.MkdirAll(filepath.Dir(innerPath), 0o770); err != nil {
					return fmt.Errorf("sink: file: file_path %s: %v", fpath, err)
				}

				f, err := os.Create(innerPath)
				if err != nil {
					return fmt.Errorf("sink: file: file_path %s: %v", innerPath, err)
				}

				if files[innerPath], err = NewFileWrapper(f, s.FileFormat, s.FileCompression); err != nil {
					return fmt.Errorf("sink: file: file_path %s: %v", innerPath, err)
				}

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
			"format", s.FileFormat.Type,
		).WithField(
			"compression", s.FileCompression.Type,
		).Debug("wrote data to file")
	}

	return nil
}
