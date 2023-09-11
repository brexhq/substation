package transform

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

func newCompressToGzip(_ context.Context, cfg config.Config) (*compressToGzip, error) {
	conf := compressGzipConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: new_compress_to_gzip: %v", err)
	}

	tf := compressToGzip{
		conf:     conf,
		isObject: false,
	}

	return &tf, nil
}

type compressToGzip struct {
	conf     compressGzipConfig
	isObject bool
}

func (tf *compressToGzip) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	gz, err := compToGzip(msg.Data())
	if err != nil {
		return nil, fmt.Errorf("transform: compress_to_gzip: %v", err)
	}

	msg.SetData(gz)
	return []*message.Message{msg}, nil
}

func (tf *compressToGzip) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
