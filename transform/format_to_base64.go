package transform

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	ibase64 "github.com/brexhq/substation/internal/base64"
	"github.com/brexhq/substation/message"
)

func newFormatToBase64(_ context.Context, cfg config.Config) (*formatToBase64, error) {
	conf := formatBase64Config{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: new_fmt_to_base64: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: new_fmt_to_base64: %v", err)
	}

	tf := formatToBase64{
		conf:     conf,
		isObject: conf.Object.Key != "" && conf.Object.SetKey != "",
	}

	return &tf, nil
}

type formatToBase64 struct {
	conf     formatBase64Config
	isObject bool
}

func (tf *formatToBase64) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	if !tf.isObject {
		b64 := ibase64.Encode(msg.Data())
		msg.SetData(b64)

		return []*message.Message{msg}, nil
	}

	value := msg.GetValue(tf.conf.Object.Key)
	b64 := ibase64.Encode(value.Bytes())

	if err := msg.SetValue(tf.conf.Object.SetKey, b64); err != nil {
		return nil, fmt.Errorf("transform: fmt_to_base64: %v", err)
	}

	return []*message.Message{msg}, nil
}

func (tf *formatToBase64) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}

func (*formatToBase64) Close(context.Context) error {
	return nil
}
