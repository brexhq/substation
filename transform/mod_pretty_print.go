package transform

import (
	"bytes"
	"context"
	gojson "encoding/json"
	"fmt"

	"golang.org/x/exp/slices"

	"github.com/brexhq/substation/config"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/json"
	"github.com/brexhq/substation/message"
)

const (
	modPPModifer           = `@this|@pretty`
	modPPOpenCurlyBracket  = 123 // {
	modPPCloseCurlyBracket = 125 // }
)

type modPrettyPrintConfig struct {
	// Direction determines whether prettyprint formatting is
	// applied or reversed.
	//
	// Must be one of:
	//
	// - to: applies prettyprint formatting
	//
	// - from: reverses prettyprint formatting
	Direction string `json:"direction"`
}

type modPrettyPrint struct {
	conf modPrettyPrintConfig

	count int
	stack []byte
}

func newModPrettyPrint(_ context.Context, cfg config.Config) (*modPrettyPrint, error) {
	conf := modPrettyPrintConfig{}
	if err := iconfig.Decode(cfg.Settings, &conf); err != nil {
		return nil, fmt.Errorf("transform: new_mod_pretty_print: %v", err)
	}

	// Validate required options.
	if conf.Direction == "" {
		return nil, fmt.Errorf("transform: new_mod_pretty_print: direction: %v", errors.ErrMissingRequiredOption)
	}

	if !slices.Contains(
		[]string{
			"to",
			"from",
		},
		conf.Direction) {
		return nil, fmt.Errorf("transform: new_mod_pretty_print: direction %q: %v", conf.Direction, errors.ErrInvalidOption)
	}

	tf := modPrettyPrint{
		conf: conf,
	}

	return &tf, nil
}

func (tf *modPrettyPrint) String() string {
	b, _ := gojson.Marshal(tf.conf)
	return string(b)
}

func (*modPrettyPrint) Close(context.Context) error {
	return nil
}

func (tf *modPrettyPrint) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	// Skip interrupt messages.
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	switch tf.conf.Direction {
	case "to":
		res := msg.GetObject(modPPModifer)

		outMsg := message.New().SetData(res.Bytes()).SetMetadata(msg.Metadata())
		return []*message.Message{outMsg}, nil
	case "from":
		for _, data := range msg.Data() {
			tf.stack = append(tf.stack, data)

			if data == modPPOpenCurlyBracket {
				tf.count++
			}

			if data == modPPCloseCurlyBracket {
				tf.count--
			}

			if tf.count == 0 {
				var buf bytes.Buffer
				if err := gojson.Compact(&buf, tf.stack); err != nil {
					return nil, fmt.Errorf("transform: proc_pretty_print: json compact: %v", err)
				}

				tf.stack = []byte{}
				if json.Valid(buf.Bytes()) {
					outMsg := message.New().SetData(buf.Bytes()).SetMetadata(msg.Metadata())
					return []*message.Message{outMsg}, nil
				}

				return nil, fmt.Errorf("transform: proc_pretty_print: invalid json")
			}
		}
	}

	return nil, nil
}
