package transform

import (
	"context"
	"encoding/json"
	"fmt"
	"unicode/utf8"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"

	ibase64 "github.com/brexhq/substation/v2/internal/base64"
)

// errFormatFromBase64DecodeBinary is returned when the Base64 transform is configured
// to decode output into an object, but the output contains binary data and
// cannot be written into a valid object.
var errFormatFromBase64DecodeBinary = fmt.Errorf("cannot write binary as object")

func newFormatFromBase64(_ context.Context, cfg config.Config) (*formatFromBase64, error) {
	conf := formatBase64Config{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform format_from_base64: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "format_from_base64"
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	tf := formatFromBase64{
		conf:     conf,
		isObject: conf.Object.SourceKey != "" && conf.Object.TargetKey != "",
	}

	return &tf, nil
}

type formatFromBase64 struct {
	conf     formatBase64Config
	isObject bool
}

func (tf *formatFromBase64) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	if !tf.isObject {
		decoded, err := ibase64.Decode(msg.Data())
		if err != nil {
			return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
		}

		msg.SetData(decoded)
		return []*message.Message{msg}, nil
	}

	value := msg.GetValue(tf.conf.Object.SourceKey)
	if !value.Exists() {
		return []*message.Message{msg}, nil
	}

	b64, err := ibase64.Decode(value.Bytes())
	if err != nil {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
	}

	if !utf8.Valid(b64) {
		return nil, errFormatFromBase64DecodeBinary
	}

	if err := msg.SetValue(tf.conf.Object.TargetKey, b64); err != nil {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
	}

	return []*message.Message{msg}, nil
}

func (tf *formatFromBase64) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
