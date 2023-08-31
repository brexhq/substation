package transform

import (
	"bytes"
	"compress/gzip"
	"context"
	gojson "encoding/json"
	"fmt"
	"io"

	"golang.org/x/exp/slices"

	"github.com/brexhq/substation/config"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/message"
)

type modGzipConfig struct {
	// Direction determines whether data is compressed or decompressed.
	//
	// Must be one of:
	//	- to (compress to gzip)
	// 	- from (decompress from gzip)
	Direction string `json:"direction"`
}

type modGzip struct {
	conf modGzipConfig
}

func newModGzip(_ context.Context, cfg config.Config) (*modGzip, error) {
	conf := modGzipConfig{}
	if err := iconfig.Decode(cfg.Settings, &conf); err != nil {
		return nil, fmt.Errorf("transform: new_mod_gzip: %v", err)
	}

	// Validate required options.
	if conf.Direction == "" {
		return nil, fmt.Errorf("transform: new_mod_gzip: type: %v", errors.ErrMissingRequiredOption)
	}

	if !slices.Contains(
		[]string{"to", "from"},
		conf.Direction) {
		return nil, fmt.Errorf("transform: new_mod_gzip: direction %q: %v", conf.Direction, errors.ErrInvalidOption)
	}

	tf := modGzip{
		conf: conf,
	}

	return &tf, nil
}

func (tf *modGzip) String() string {
	b, _ := gojson.Marshal(tf.conf)
	return string(b)
}

func (*modGzip) Close(context.Context) error {
	return nil
}

func (tf *modGzip) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	// Skip interrupt messages.
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	var value []byte
	switch tf.conf.Direction {
	case "from":
		from, err := tf.from(msg.Data())
		if err != nil {
			return nil, fmt.Errorf("transform: mod_gzip: %v", err)
		}

		value = from
	case "to":
		to, err := tf.to(msg.Data())
		if err != nil {
			return nil, fmt.Errorf("transform: mod_gzip: %v", err)
		}

		value = to
	}

	finMsg := message.New().SetData(value).SetMetadata(msg.Metadata())
	return []*message.Message{finMsg}, nil
}

func (t *modGzip) from(data []byte) ([]byte, error) {
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

func (t *modGzip) to(data []byte) ([]byte, error) {
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
