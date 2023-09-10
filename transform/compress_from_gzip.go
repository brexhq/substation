package transform

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

func newCompressFromGzip(_ context.Context, cfg config.Config) (*compressFromGzip, error) {
	conf := compressGzipConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: new_compress_from_gzip: %v", err)
	}

	tf := compressFromGzip{
		conf:     conf,
		isObject: false,
	}

	return &tf, nil
}

type compressFromGzip struct {
	conf     compressGzipConfig
	isObject bool
}

func (tf *compressFromGzip) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	gz, err := compFromGzip(msg.Data())
	if err != nil {
		return nil, fmt.Errorf("transform: compress_from_gzip: %v", err)
	}

	msg.SetData(gz)
	return []*message.Message{msg}, nil
}

func (tf *compressFromGzip) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}

func (*compressFromGzip) Close(context.Context) error {
	return nil
}
