package transform

import (
	"context"
	"encoding/json"
	"fmt"
	"unicode/utf8"

	"github.com/brexhq/substation/config"
	ibase64 "github.com/brexhq/substation/internal/base64"
	"github.com/brexhq/substation/message"
)

// errFormatFromBase64DecodeBinary is returned when the Base64 transform is configured
// to decode output into an object, but the output contains binary data and
// cannot be written into a valid object.
var errFormatFromBase64DecodeBinary = fmt.Errorf("cannot write binary as object")

func newFormatFromBase64(_ context.Context, cfg config.Config) (*formatFromBase64, error) {
	conf := formatBase64Config{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: new_caseDown: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: new_caseDown: %v", err)
	}

	tf := formatFromBase64{
		conf:     conf,
		isObject: conf.Object.Key != "" && conf.Object.SetKey != "",
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
			return nil, fmt.Errorf("transform: decode_base64: %v", err)
		}

		msg.SetData(decoded)
		return []*message.Message{msg}, nil
	}

	result := msg.GetValue(tf.conf.Object.Key)
	decoded, err := ibase64.Decode(result.Bytes())
	if err != nil {
		return nil, fmt.Errorf("transform: decode_base64: %v", err)
	}

	if !utf8.Valid(decoded) {
		return nil, errFormatFromBase64DecodeBinary
	}

	if err := msg.SetValue(tf.conf.Object.SetKey, decoded); err != nil {
		return nil, fmt.Errorf("transform: decode_base64: %v", err)
	}

	return []*message.Message{msg}, nil
}

func (tf *formatFromBase64) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
