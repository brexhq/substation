package transform

import (
	"context"
	"encoding/json"
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
	Buffer iconfig.Buffer `json:"buffer"`

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

func (c *sendFileConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *sendFileConfig) Validate() error {
	if c.FileFormat.Type == "" {
		c.FileFormat.Type = "json"
	}

	if c.FileCompression.Type == "" {
		c.FileCompression.Type = "gzip"
	}

	return nil
}

func newSendFile(_ context.Context, cfg config.Config) (*sendFile, error) {
	conf := sendFileConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: send_file: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: send_file: %v", err)
	}

	tf := sendFile{
		conf: conf,
	}

	// File extensions are dynamic and not directly configurable.
	tf.extension = file.NewExtension(conf.FileFormat, conf.FileCompression)

	buffer, err := aggregate.New(aggregate.Config{
		Count:    conf.Buffer.Count,
		Size:     conf.Buffer.Size,
		Duration: conf.Buffer.Duration,
	})
	if err != nil {
		return nil, fmt.Errorf("transform: send_file: %v", err)
	}
	tf.buffer = buffer

	return &tf, nil
}

type sendFile struct {
	conf sendFileConfig

	extension string

	mu     sync.Mutex
	buffer *aggregate.Aggregate
}

func (tf *sendFile) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	tf.mu.Lock()
	defer tf.mu.Unlock()

	if msg.IsControl() {
		for prefix := range tf.buffer.GetAll() {
			if err := tf.writeFile(prefix); err != nil {
				return nil, fmt.Errorf("transform: send_file: prefix %s: %v", prefix, err)
			}
		}

		tf.buffer.ResetAll()
		return []*message.Message{msg}, nil
	}

	prefix := msg.GetValue(tf.conf.FilePath.PrefixKey).String()
	// Writes data as a file only when the buffer is full.
	if ok := tf.buffer.Add(prefix, msg.Data()); ok {
		return []*message.Message{msg}, nil
	}

	if err := tf.writeFile(prefix); err != nil {
		return nil, fmt.Errorf("transform: send_file: %v", err)
	}

	// Reset the buffer and add the msg data.
	tf.buffer.Reset(prefix)
	_ = tf.buffer.Add(prefix, msg.Data())

	return []*message.Message{msg}, nil
}

func (t *sendFile) writeFile(prefix string) error {
	if t.buffer.Count(prefix) == 0 {
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

	for _, rec := range t.buffer.Get(prefix) {
		if _, err := w.Write(rec); err != nil {
			return err
		}
	}

	if err := w.Close(); err != nil {
		return err
	}

	return nil
}

func (tf *sendFile) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
