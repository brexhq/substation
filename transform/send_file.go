package transform

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/brexhq/substation/config"
	_config "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/file"
	mess "github.com/brexhq/substation/message"
	"github.com/google/uuid"
)

type sendFileConfig struct {
	// FilePath determines how the name of the file is constructed.
	// See filePath.New for more information.
	FilePath file.Path `json:"file_path"`
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

type sendFile struct {
	conf sendFileConfig

	path      string
	extension string
	mu        *sync.Mutex
	buffer    map[string]*file.Wrapper
}

func newSendFile(_ context.Context, cfg config.Config) (*sendFile, error) {
	conf := sendFileConfig{}
	if err := _config.Decode(cfg.Settings, &conf); err != nil {
		return nil, err
	}

	if conf.FileFormat.Type == "" {
		conf.FileFormat.Type = "json"
	}

	if conf.FileCompression.Type == "" {
		conf.FileCompression.Type = "gzip"
	}

	send := sendFile{
		conf: conf,
	}

	// File extensions are dynamic and not directly configurable.
	send.extension = file.NewExtension(conf.FileFormat, conf.FileCompression)
	now := time.Now()

	// The default file path is: cwd/year/month/day/uuid.extension.
	send.path = conf.FilePath.New()
	if send.path == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("send: file: %v", err)
		}

		send.path = path.Join(
			cwd,
			now.Format("2006"), now.Format("01"), now.Format("02"),
			uuid.New().String(),
		) + send.extension
	} else if conf.FilePath.Extension {
		send.path += send.extension
	}

	// Ensures that the path is OS agnostic.
	send.path = filepath.FromSlash(send.path)

	send.mu = &sync.Mutex{}
	send.buffer = make(map[string]*file.Wrapper)

	return &send, nil
}

func (t *sendFile) Close(context.Context) error {
	// Lock the transform to prevent concurrent access to the buffer.
	t.mu.Lock()
	defer t.mu.Unlock()

	for _, f := range t.buffer {
		f.Close()
	}

	return nil
}

func (t *sendFile) Transform(ctx context.Context, messages ...*mess.Message) ([]*mess.Message, error) {
	// Lock the transform to prevent concurrent access to the buffer.
	t.mu.Lock()
	defer t.mu.Unlock()

	control := false
	for _, message := range messages {
		if message.IsControl() {
			control = true
			continue
		}

		// path is used so that key values can be interpolated into the file path.
		path := t.path

		// If either prefix or suffix keys are set, then the object name is non-default
		// and can be safely interpolated. If either are empty strings, then an error
		// is returned.
		if t.conf.FilePath.PrefixKey != "" {
			prefix := message.Get(t.conf.FilePath.PrefixKey).String()
			if prefix == "" {
				return nil, fmt.Errorf("send: file: %v", fmt.Errorf("empty prefix string"))
			}

			path = strings.Replace(path, "${PATH_PREFIX}", prefix, 1)
		}
		if t.conf.FilePath.SuffixKey != "" {
			suffix := message.Get(t.conf.FilePath.SuffixKey).String()
			if suffix == "" {
				return nil, fmt.Errorf("send: file: %v", fmt.Errorf("empty suffix string"))
			}

			path = strings.Replace(path, "${PATH_SUFFIX}", suffix, 1)
		}

		if _, ok := t.buffer[path]; !ok {
			if err := os.MkdirAll(filepath.Dir(path), 0o770); err != nil {
				return nil, fmt.Errorf("send: file: file_path %s: %v", t.path, err)
			}

			f, err := os.Create(path)
			if err != nil {
				return nil, fmt.Errorf("send: file: file_path %s: %v", path, err)
			}

			if t.buffer[path], err = file.NewWrapper(f, t.conf.FileFormat, t.conf.FileCompression); err != nil {
				return nil, fmt.Errorf("send: file: file_path %s: %v", path, err)
			}
		}

		if _, err := t.buffer[path].Write(message.Data()); err != nil {
			return nil, fmt.Errorf("send: file: file_path %s: %v", path, err)
		}
	}

	// If a control message is received, then files are closed and removed from the
	// buffer.
	if !control {
		return messages, nil
	}

	for path := range t.buffer {
		t.buffer[path].Close()
		delete(t.buffer, path)
	}

	return messages, nil
}
