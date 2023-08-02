package transform

import (
	"bytes"
	"context"
	gojson "encoding/json"
	"fmt"

	"golang.org/x/exp/slices"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/json"
	mess "github.com/brexhq/substation/message"
)

const (
	// used with json.Get, returns a pretty printed root JSON object
	procPPModifer           = `@this|@pretty`
	procPPOpenCurlyBracket  = 123 // {
	procPPCloseCurlyBracket = 125 // }
)

type procPrettyPrintConfig struct {
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

type procPrettyPrint struct {
	conf procPrettyPrintConfig

	count int
	stack []byte
}

func newProcPrettyPrint(_ context.Context, cfg config.Config) (*procPrettyPrint, error) {
	conf := procPrettyPrintConfig{}
	if err := config.Decode(cfg.Settings, &conf); err != nil {
		return nil, err
	}

	// Validate required options.
	if conf.Direction == "" {
		return nil, fmt.Errorf("transform: proc_pretty_print: direction: %v", errors.ErrMissingRequiredOption)
	}

	if !slices.Contains(
		[]string{
			"to",
			"from",
		},
		conf.Direction) {
		return nil, fmt.Errorf("transform: proc_pretty_print: direction %q: %v", conf.Direction, errors.ErrInvalidOption)
	}

	proc := procPrettyPrint{
		conf: conf,
	}

	return &proc, nil
}

func (t *procPrettyPrint) String() string {
	b, _ := gojson.Marshal(t.conf)
	return string(b)
}

func (*procPrettyPrint) Close(context.Context) error {
	return nil
}

func (t *procPrettyPrint) Transform(ctx context.Context, messages ...*mess.Message) ([]*mess.Message, error) {
	var output []*mess.Message

	for _, message := range messages {
		// Skip control messages.
		if message.IsControl() {
			output = append(output, message)
			continue
		}

		switch t.conf.Direction {
		case "to":
			res := message.Get(procPPModifer).String()
			msg, err := mess.New(
				mess.SetData([]byte(res)),
				mess.SetMetadata(message.Metadata()),
			)
			if err != nil {
				return nil, fmt.Errorf("process: dns: %v", err)
			}

			output = append(output, msg)
		case "from":
			for _, data := range message.Data() {
				t.stack = append(t.stack, data)

				if data == procPPOpenCurlyBracket {
					t.count++
				}

				if data == procPPCloseCurlyBracket {
					t.count--
				}

				if t.count == 0 {
					var buf bytes.Buffer
					if err := gojson.Compact(&buf, t.stack); err != nil {
						return nil, fmt.Errorf("transform: proc_pretty_print: json compact: %v", err)
					}

					if json.Valid(buf.Bytes()) {
						msg, err := mess.New(
							mess.SetData(buf.Bytes()),
							mess.SetMetadata(message.Metadata()),
						)
						if err != nil {
							return nil, fmt.Errorf("transform: proc_pretty_print: %v", err)
						}

						output = append(output, msg)
					}

					t.stack = []byte{}
				}
			}
		}
	}

	return output, nil
}
