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
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/file"
	"github.com/brexhq/substation/message"
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
	if err := iconfig.Decode(cfg.Settings, &conf); err != nil {
		return nil, fmt.Errorf("transform: new_send_file: %v", err)
	}

	tf := sendFile{
		conf: conf,
	}

	// File extensions are dynamic and not directly configurable.
	tf.extension = file.NewExtension(conf.FileFormat, conf.FileCompression)

	tf.mu = sync.Mutex{}
	tf.buffer = make(map[string]*aggregate.Aggregate)
	tf.bufferCfg = aggregate.Config{
		Count:    conf.Buffer.Count,
		Size:     conf.Buffer.Size,
		Duration: conf.Buffer.Duration,
	}

	return &tf, nil
}

func (*sendFile) Close(context.Context) error {
	return nil
}

func (tf *sendFile) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	// Lock the transform to prevent concurrent access to the buffer.
	tf.mu.Lock()
	defer tf.mu.Unlock()

	if msg.IsControl() {
		for path := range tf.buffer {
			if err := tf.writeFile(path); err != nil {
				return nil, fmt.Errorf("transform: send_file: file_path %s: %v", path, err)
			}
		}

		tf.buffer = make(map[string]*aggregate.Aggregate)
		return []*message.Message{msg}, nil
	}

	var prefixKey string
	if tf.conf.FilePath.PrefixKey != "" {
		prefixKey = msg.GetObject(tf.conf.FilePath.PrefixKey).String()
	}

	if _, ok := tf.buffer[prefixKey]; !ok {
		agg, err := aggregate.New(tf.bufferCfg)
		if err != nil {
			return nil, fmt.Errorf("transform: send_file: %v", err)
		}

		tf.buffer[prefixKey] = agg
	}

	// Writes data as a file only when the buffer is full.
	if ok := tf.buffer[prefixKey].Add(msg.Data()); ok {
		return []*message.Message{msg}, nil
	}

	if err := tf.writeFile(prefixKey); err != nil {
		return nil, fmt.Errorf("transform: send_file: %v", err)
	}

	// Reset the buffer and add the msg data.
	tf.buffer[prefixKey].Reset()
	_ = tf.buffer[prefixKey].Add(msg.Data())

	return []*message.Message{msg}, nil
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
