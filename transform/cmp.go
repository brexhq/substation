package transform

import (
	"bytes"
	"compress/gzip"
	"io"

	iconfig "github.com/brexhq/substation/internal/config"
)

type cmpGzipConfig struct{}

func (c *cmpGzipConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func cmpGzipCompress(data []byte) ([]byte, error) {
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

func cmpGzipDecompress(data []byte) ([]byte, error) {
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
