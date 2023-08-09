package transform

import (
	"context"
	gojson "encoding/json"
	"fmt"
	"unicode/utf8"

	"golang.org/x/exp/slices"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/base64"
	_config "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	mess "github.com/brexhq/substation/message"
)

// errProcBase64DecodedBinary is returned when the Base64 transform is configured
// to decode output into an object, but the output contains binary data and
// cannot be written into a valid object.
var errProcBase64DecodedBinary = fmt.Errorf("cannot write binary as object")

type procBase64Config struct {
	// Key retrieves a value from an object for processing.
	//
	// This is optional for transforms that support processing non-object data.
	Key string `json:"key"`
	// SetKey inserts a processed value into an object.
	//
	// This is optional for transforms that support processing non-object data.
	SetKey string `json:"set_key"`
	// Direction determines whether data is encoded or decoded.
	//
	// Must be one of:
	//
	// - to: encode to base64
	//
	// - from: decode from base64
	Direction string `json:"direction"`
}

type procBase64 struct {
	conf     procBase64Config
	isObject bool
}

func newProcBase64(_ context.Context, cfg config.Config) (*procBase64, error) {
	conf := procBase64Config{}
	if err := _config.Decode(cfg.Settings, &conf); err != nil {
		return nil, err
	}

	// Validate required options.
	if (conf.Key == "" && conf.SetKey != "") || (conf.Key != "" && conf.SetKey == "") {
		return nil, fmt.Errorf("transform: proc_base64: key %s set_key %s: %v", conf.Key, conf.SetKey, errInvalidDataPattern)
	}

	if conf.Direction == "" {
		return nil, fmt.Errorf("transform: proc_base64: direction: %v", errors.ErrMissingRequiredOption)
	}

	if !slices.Contains(
		[]string{
			"to",
			"from",
		},
		conf.Direction) {
		return nil, fmt.Errorf("transform: proc_base64: direction %s: %v", conf.Direction, errors.ErrInvalidOption)
	}

	proc := procBase64{
		conf:     conf,
		isObject: conf.Key != "" && conf.SetKey != "",
	}

	return &proc, nil
}

func (t *procBase64) String() string {
	b, _ := gojson.Marshal(t.conf)
	return string(b)
}

func (*procBase64) Close(context.Context) error {
	return nil
}

func (t *procBase64) Transform(ctx context.Context, messages ...*mess.Message) ([]*mess.Message, error) {
	var output []*mess.Message

	for _, message := range messages {
		// Skip control messages.
		if message.IsControl() {
			output = append(output, message)
			continue
		}

		switch t.isObject {
		case true:
			result := message.Get(t.conf.Key).String()
			tmp := []byte(result)

			var value []byte
			switch t.conf.Direction {
			case "from":
				decode, err := base64.Decode(tmp)
				if err != nil {
					return nil, fmt.Errorf("transform: proc_base64: %v", err)
				}

				if !utf8.Valid(decode) {
					return nil, fmt.Errorf("transform: proc_base64: %v", errProcBase64DecodedBinary)
				}

				value = decode
			case "to":
				value = base64.Encode(tmp)
			}

			if err := message.Set(t.conf.SetKey, value); err != nil {
				return nil, fmt.Errorf("transform: proc_base64: %v", err)
			}

			output = append(output, message)
		case false:
			var value []byte
			switch t.conf.Direction {
			case "from":
				decode, err := base64.Decode(message.Data())
				if err != nil {
					return nil, fmt.Errorf("transform: proc_base64: %v", err)
				}

				value = decode
			case "to":
				value = base64.Encode(message.Data())
			}

			msg, err := mess.New(
				mess.SetData(value),
				mess.SetMetadata(message.Metadata()),
			)
			if err != nil {
				return nil, fmt.Errorf("transform: proc_base64: %v", err)
			}

			output = append(output, msg)
		}
	}

	return output, nil
}
