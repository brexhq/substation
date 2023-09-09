package transform

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

func newCmpFromGzip(_ context.Context, cfg config.Config) (*cmpFromGzip, error) {
	conf := cmpGzipConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: new_compress_from_gzip: %v", err)
	}

	tf := cmpFromGzip{
		conf:     conf,
		isObject: false,
	}

	return &tf, nil
}

type cmpFromGzip struct {
	conf     cmpGzipConfig
	isObject bool
}

func (tf *cmpFromGzip) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	// Skip interrupt messages.
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	gz, err := cmpGzipDecompress(msg.Data())
	if err != nil {
		return nil, fmt.Errorf("transform: compress_from_gzip: %v", err)
	}

	msg.SetData(gz)
	return []*message.Message{msg}, nil
}

func (tf *cmpFromGzip) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}

func (*cmpFromGzip) Close(context.Context) error {
	return nil
}
