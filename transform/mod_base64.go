package transform

import (
	"context"
	gojson "encoding/json"
	"fmt"
	"unicode/utf8"

	"golang.org/x/exp/slices"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/base64"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/message"
)

// errModBase64DecodedBinary is returned when the Base64 transform is configured
// to decode output into an object, but the output contains binary data and
// cannot be written into a valid object.
var errModBase64DecodedBinary = fmt.Errorf("cannot write binary as object")

type modBase64Config struct {
	Object configObject `json:"object"`

	// Direction determines whether data is encoded or decoded.
	//
	// Must be one of:
	//	- to (encode to base64)
	//	- from (decode from base64)
	Direction string `json:"direction"`
}

type modBase64 struct {
	conf  modBase64Config
	isObj bool
}

func newModBase64(_ context.Context, cfg config.Config) (*modBase64, error) {
	conf := modBase64Config{}
	if err := iconfig.Decode(cfg.Settings, &conf); err != nil {
		return nil, fmt.Errorf("transform: new_mod_base64: %v", err)
	}

	// Validate required options.
	if conf.Object.Key == "" && conf.Object.SetKey != "" {
		return nil, fmt.Errorf("transform: new_mod_base64: object_key: %v", errors.ErrMissingRequiredOption)
	}

	if conf.Object.Key != "" && conf.Object.SetKey == "" {
		return nil, fmt.Errorf("transform: new_mod_base64: object_set_key: %v", errors.ErrMissingRequiredOption)
	}

	if conf.Direction == "" {
		return nil, fmt.Errorf("transform: new_mod_base64: direction: %v", errors.ErrMissingRequiredOption)
	}

	if !slices.Contains(
		[]string{
			"to",
			"from",
		},
		conf.Direction) {
		return nil, fmt.Errorf("transform: new_mod_base64: direction %s: %v", conf.Direction, errors.ErrInvalidOption)
	}

	tf := modBase64{
		conf:  conf,
		isObj: conf.Object.Key != "" && conf.Object.SetKey != "",
	}

	return &tf, nil
}

func (tf *modBase64) String() string {
	b, _ := gojson.Marshal(tf.conf)
	return string(b)
}

func (*modBase64) Close(context.Context) error {
	return nil
}

func (tf *modBase64) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	// Skip interrupt messages.
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	if !tf.isObj {
		var val []byte
		switch tf.conf.Direction {
		case "from":
			v, err := tf.decode(msg.Data())
			if err != nil {
				return nil, fmt.Errorf("transform: mod_base64: %v", err)
			}

			val = v
		case "to":
			val = base64.Encode(msg.Data())
		}

		outMsg := message.New().SetData(val).SetMetadata(msg.Metadata())
		return []*message.Message{outMsg}, nil
	}

	result := msg.GetObject(tf.conf.Object.Key).Bytes()

	var val []byte
	switch tf.conf.Direction {
	case "from":
		v, err := tf.decode(result)
		if err != nil {
			return nil, fmt.Errorf("transform: mod_base64: %v", err)
		}

		val = v
	case "to":
		val = base64.Encode(result)
	}

	if err := msg.SetObject(tf.conf.Object.SetKey, val); err != nil {
		return nil, fmt.Errorf("transform: mod_base64: %v", err)
	}

	return []*message.Message{msg}, nil
}

func (tf *modBase64) decode(b []byte) ([]byte, error) {
	decode, err := base64.Decode(b)
	if err != nil {
		return nil, err
	}

	if tf.isObj && !utf8.Valid(decode) {
		return nil, errModBase64DecodedBinary
	}

	return decode, nil
}
