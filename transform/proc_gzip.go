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
	_config "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	mess "github.com/brexhq/substation/message"
)

type procGzipConfig struct {
	// Key retrieves a value from an object for processing.
	//
	// This is optional for transforms that support processing non-object data.
	Key string `json:"key"`
	// SetKey inserts a processed value into an object.
	//
	// This is optional for transforms that support processing non-object data.
	SetKey string `json:"set_key"`
	// Direction determines whether data is compressed or decompressed.
	//
	// Must be one of:
	//	- to: compress to gzip
	// 	- from: decompress from gzip
	Direction string `json:"direction"`
}

type procGzip struct {
	conf procGzipConfig
}

func newProcGzip(_ context.Context, cfg config.Config) (*procGzip, error) {
	conf := procGzipConfig{}
	if err := _config.Decode(cfg.Settings, &conf); err != nil {
		return nil, err
	}

	// Validate required options.
	if (conf.Key != "" && conf.SetKey == "") ||
		(conf.Key == "" && conf.SetKey != "") {
		return nil, fmt.Errorf("transform: proc_gzip: key %s set_key %s: %v", conf.Key, conf.SetKey, errInvalidDataPattern)
	}

	if conf.Direction == "" {
		return nil, fmt.Errorf("transform: proc_gzip: type: %v", errors.ErrMissingRequiredOption)
	}

	if !slices.Contains(
		[]string{"to", "from"},
		conf.Direction) {
		return nil, fmt.Errorf("transform: proc_gzip: direction %q: %v", conf.Direction, errors.ErrInvalidOption)
	}

	proc := procGzip{
		conf: conf,
	}

	return &proc, nil
}

func (t *procGzip) String() string {
	b, _ := gojson.Marshal(t.conf)
	return string(b)
}

func (*procGzip) Close(context.Context) error {
	return nil
}

func (t *procGzip) Transform(ctx context.Context, messages ...*mess.Message) ([]*mess.Message, error) {
	var output []*mess.Message

	for _, message := range messages {
		// Skip control messages.
		if message.IsControl() {
			output = append(output, message)
			continue
		}

		var value []byte
		switch t.conf.Direction {
		case "from":
			from, err := t.from(message.Data())
			if err != nil {
				return nil, fmt.Errorf("transform: proc_gzip: %v", err)
			}

			value = from
		case "to":
			to, err := t.to(message.Data())
			if err != nil {
				return nil, fmt.Errorf("transform: proc_gzip: %v", err)
			}

			value = to
		}

		msg, err := mess.New(
			mess.SetData(value),
			mess.SetMetadata(message.Metadata()),
		)
		if err != nil {
			return nil, fmt.Errorf("transform: proc_gzip: %v", err)
		}

		output = append(output, msg)
	}

	return output, nil
}

func (t *procGzip) from(data []byte) ([]byte, error) {
	r := bytes.NewReader(data)
	gz, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("transform: proc_gzip: %v", err)
	}

	output, err := io.ReadAll(gz)
	if err != nil {
		return nil, fmt.Errorf("transform: proc_gzip: %v", err)
	}

	return output, nil
}

func (t *procGzip) to(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(data); err != nil {
		return nil, fmt.Errorf("transform: proc_gzip: %v", err)
	}
	if err := gz.Close(); err != nil {
		return nil, fmt.Errorf("transform: proc_gzip: %v", err)
	}

	return buf.Bytes(), nil
}
