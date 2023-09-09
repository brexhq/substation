package transform

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/message"
)

const (
	fmtFromPPOpenCurlyBracket  = 123 // {
	fmtFromPPCloseCurlyBracket = 125 // }
)

type fmtFromPrettyPrintConfig struct{}

func (c *fmtFromPrettyPrintConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

type fmtFromPrettyPrint struct {
	conf fmtFromPrettyPrintConfig

	count int
	stack []byte
}

func newfmtFromPrettyPrint(_ context.Context, cfg config.Config) (*fmtFromPrettyPrint, error) {
	conf := fmtFromPrettyPrintConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: new_fmt_from_pretty_print: %v", err)
	}

	tf := fmtFromPrettyPrint{
		conf: conf,
	}

	return &tf, nil
}

func (tf *fmtFromPrettyPrint) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	// Skip interrupt messages.
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	for _, data := range msg.Data() {
		tf.stack = append(tf.stack, data)

		if data == fmtFromPPOpenCurlyBracket {
			tf.count++
		}

		if data == fmtFromPPCloseCurlyBracket {
			tf.count--
		}

		if tf.count == 0 {
			var buf bytes.Buffer
			if err := json.Compact(&buf, tf.stack); err != nil {
				return nil, fmt.Errorf("transform: fmt_from_pretty_print: json compact: %v", err)
			}

			tf.stack = []byte{}
			if json.Valid(buf.Bytes()) {
				msg.SetData(buf.Bytes())
				return []*message.Message{msg}, nil
			}

			return nil, fmt.Errorf("transform: fmt_from_pretty_print: invalid json")
		}
	}

	return nil, nil
}

func (tf *fmtFromPrettyPrint) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}

func (*fmtFromPrettyPrint) Close(context.Context) error {
	return nil
}
