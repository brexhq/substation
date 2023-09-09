package transform

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

func newCmpToGzip(_ context.Context, cfg config.Config) (*cmpToGzip, error) {
	conf := cmpGzipConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: new_compress_to_gzip: %v", err)
	}

	tf := cmpToGzip{
		conf:     conf,
		isObject: false,
	}

	return &tf, nil
}

type cmpToGzip struct {
	conf     cmpGzipConfig
	isObject bool
}

func (tf *cmpToGzip) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	// Skip interrupt messages.
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	gz, err := cmpGzipCompress(msg.Data())
	if err != nil {
		return nil, fmt.Errorf("transform: compress_to_gzip: %v", err)
	}

	msg.SetData(gz)
	return []*message.Message{msg}, nil
}

func (tf *cmpToGzip) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}

func (*cmpToGzip) Close(context.Context) error {
	return nil
}
