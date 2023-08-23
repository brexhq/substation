package transform

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/aggregate"
	_config "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/file"
	mess "github.com/brexhq/substation/message"
)

type sendFileConfig struct {
	Buffer aggregate.Config `json:"buffer"`
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

	extension string
	// buffer is safe for concurrent use.
	mu        sync.Mutex
	buffer    map[string]*aggregate.Aggregate
	bufferCfg aggregate.Config
}

func newSendFile(_ context.Context, cfg config.Config) (*sendFile, error) {
	conf := sendFileConfig{}
	if err := _config.Decode(cfg.Settings, &conf); err != nil {
		return nil, err
	}

	send := sendFile{
		conf: conf,
	}

	// File extensions are dynamic and not directly configurable.
	send.extension = file.NewExtension(conf.FileFormat, conf.FileCompression)

	send.mu = sync.Mutex{}
	send.buffer = make(map[string]*aggregate.Aggregate)
	send.bufferCfg = aggregate.Config{
		Count:    conf.Buffer.Count,
		Size:     conf.Buffer.Size,
		Interval: conf.Buffer.Interval,
	}

	return &send, nil
}

func (*sendFile) Close(context.Context) error {
	return nil
}

func (send *sendFile) Transform(ctx context.Context, message *mess.Message) ([]*mess.Message, error) {
	// Lock the transform to prevent concurrent access to the buffer.
	send.mu.Lock()
	defer send.mu.Unlock()

	if message.IsControl() {
		for path := range send.buffer {
			if err := send.writeFile(path); err != nil {
				return nil, fmt.Errorf("transform: send_file: file_path %s: %v", path, err)
			}
		}

		send.buffer = make(map[string]*aggregate.Aggregate)
		return []*mess.Message{message}, nil
	}

	var prefixKey string
	if send.conf.FilePath.PrefixKey != "" {
		prefixKey = message.Get(send.conf.FilePath.PrefixKey).String()
	}

	if _, ok := send.buffer[prefixKey]; !ok {
		agg, err := aggregate.New(send.bufferCfg)
		if err != nil {
			return nil, fmt.Errorf("transform: send_file: %v", err)
		}

		send.buffer[prefixKey] = agg
	}

	// Writes data as a file only when the buffer is full.
	if ok := send.buffer[prefixKey].Add(message.Data()); ok {
		return []*mess.Message{message}, nil
	}

	if err := send.writeFile(prefixKey); err != nil {
		return nil, fmt.Errorf("transform: send_file: %v", err)
	}

	// Reset the buffer and add the message data.
	send.buffer[prefixKey].Reset()
	_ = send.buffer[prefixKey].Add(message.Data())

	return []*mess.Message{message}, nil
}

func (t *sendFile) writeFile(prefix string) error {
	if t.buffer[prefix].Count() == 0 {
		return nil
	}

	fpath := t.conf.FilePath.New()
	if fpath == "" {
		return fmt.Errorf("file_path is empty")
	}

	if prefix != "" {
		fpath = strings.Replace(fpath, "${PATH_PREFIX}", prefix, 1)
	}

	fpath += t.extension

	// Ensures that the path is OS agnostic.
	fpath = filepath.FromSlash(fpath)
	if err := os.MkdirAll(filepath.Dir(fpath), 0o770); err != nil {
		return err
	}

	f, err := os.Create(fpath)
	if err != nil {
		fmt.Println(fpath)
		return err
	}

	w, err := file.NewWrapper(f, t.conf.FileFormat, t.conf.FileCompression)
	if err != nil {
		return err
	}

	for _, rec := range t.buffer[prefix].Get() {
		if _, err := w.Write(rec); err != nil {
			return err
		}
	}

	if err := w.Close(); err != nil {
		return err
	}

	return nil
}
