package transform

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

func newFormatToGzip(_ context.Context, cfg config.Config) (*formatToGzip, error) {
	conf := formatGzipConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform format_to_gzip: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "format_to_gzip"
	}

	tf := formatToGzip{
		conf:     conf,
		isObject: false,
	}

	return &tf, nil
}

type formatToGzip struct {
	conf     formatGzipConfig
	isObject bool
}

func (tf *formatToGzip) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	gz, err := fmtToGzip(msg.Data())
	if err != nil {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
	}

	msg.SetData(gz)
	return []*message.Message{msg}, nil
}

func (tf *formatToGzip) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
