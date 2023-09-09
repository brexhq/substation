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

// errFmtFromBase64DecodeBinary is returned when the Base64 transform is configured
// to decode output into an object, but the output contains binary data and
// cannot be written into a valid object.
var errFmtFromBase64DecodeBinary = fmt.Errorf("cannot write binary as object")

func newFmtFromBase64(_ context.Context, cfg config.Config) (*fmtFromBase64, error) {
	conf := fmtBase64Config{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: new_caseDown: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: new_caseDown: %v", err)
	}

	tf := fmtFromBase64{
		conf:     conf,
		isObject: conf.Object.Key != "" && conf.Object.SetKey != "",
	}

	return &tf, nil
}

type fmtFromBase64 struct {
	conf     fmtBase64Config
	isObject bool
}

func (tf *fmtFromBase64) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	// Skip control messages.
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
		return nil, errFmtFromBase64DecodeBinary
	}

	if err := msg.SetValue(tf.conf.Object.SetKey, decoded); err != nil {
		return nil, fmt.Errorf("transform: decode_base64: %v", err)
	}

	return []*message.Message{msg}, nil
}

func (tf *fmtFromBase64) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}

func (*fmtFromBase64) Close(context.Context) error {
	return nil
}
