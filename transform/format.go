package transform

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"

	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
)

type formatBase64Config struct {
	ID     string         `json:"id"`
	Object iconfig.Object `json:"object"`
}

func (c *formatBase64Config) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *formatBase64Config) Validate() error {
	if c.Object.SourceKey == "" && c.Object.TargetKey != "" {
		return fmt.Errorf("object_source_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.Object.SourceKey != "" && c.Object.TargetKey == "" {
		return fmt.Errorf("object_target_key: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

type formatGzipConfig struct {
	ID string `json:"id"`
}

func (c *formatGzipConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func fmtToGzip(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(data); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func fmtFromGzip(data []byte) ([]byte, error) {
	r := bytes.NewReader(data)
	gz, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	output, err := io.ReadAll(gz)
	if err != nil {
		return nil, err
	}

	return output, nil
}
