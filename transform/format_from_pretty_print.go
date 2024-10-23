package transform

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"

	iconfig "github.com/brexhq/substation/v2/internal/config"
)

const (
	formatFromPPOpenCurlyBracket  = 123 // {
	formatFromPPCloseCurlyBracket = 125 // }
)

type formatFromPrettyPrintConfig struct {
	ID string `json:"id"`
}

func (c *formatFromPrettyPrintConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func newFormatFromPrettyPrint(_ context.Context, cfg config.Config) (*formatFromPrettyPrint, error) {
	conf := formatFromPrettyPrintConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform format_from_pretty_print: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "format_from_pretty_print"
	}

	tf := formatFromPrettyPrint{
		conf: conf,
	}

	return &tf, nil
}

type formatFromPrettyPrint struct {
	conf formatFromPrettyPrintConfig

	count int
	stack []byte
}

func (tf *formatFromPrettyPrint) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.HasFlag(message.IsControl) {
		return []*message.Message{msg}, nil
	}

	for _, data := range msg.Data() {
		tf.stack = append(tf.stack, data)

		if data == formatFromPPOpenCurlyBracket {
			tf.count++
		}

		if data == formatFromPPCloseCurlyBracket {
			tf.count--
		}

		if tf.count == 0 {
			var buf bytes.Buffer
			if err := json.Compact(&buf, tf.stack); err != nil {
				return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
			}

			tf.stack = []byte{}
			if json.Valid(buf.Bytes()) {
				msg.SetData(buf.Bytes())
				return []*message.Message{msg}, nil
			}

			return nil, fmt.Errorf("transform %s: invalid json", tf.conf.ID)
		}
	}

	return nil, nil
}

func (tf *formatFromPrettyPrint) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
