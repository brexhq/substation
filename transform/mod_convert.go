package transform

import (
	"context"
	gojson "encoding/json"
	"fmt"

	"golang.org/x/exp/slices"

	"github.com/brexhq/substation/config"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/message"
)

type modConvertConfig struct {
	Object configObject `json:"object"`

	// Type is the target conversion type.
	//
	// Must be one of:
	//	- bool (boolean)
	//	- int (integer)
	//	- float
	//	- uint (unsigned integer)
	//	- string
	Type string `json:"type"`
}

type modConvert struct {
	conf modConvertConfig
}

func newModConvert(_ context.Context, cfg config.Config) (*modConvert, error) {
	conf := modConvertConfig{}
	if err := iconfig.Decode(cfg.Settings, &conf); err != nil {
		return nil, fmt.Errorf("transform: new_mod_convert: %v", err)
	}

	// Validate required options.
	if conf.Object.Key == "" {
		return nil, fmt.Errorf("transform: new_mod_convert: object_key: %v", errors.ErrMissingRequiredOption)
	}

	if conf.Object.SetKey == "" {
		return nil, fmt.Errorf("transform: new_mod_convert: object_set_key: %v", errors.ErrMissingRequiredOption)
	}

	if conf.Type == "" {
		return nil, fmt.Errorf("transform: new_mod_convert: type: %v", errors.ErrMissingRequiredOption)
	}

	if !slices.Contains(
		[]string{
			"bool",
			"int",
			"float",
			"uint",
			"string",
		},
		conf.Type) {
		return nil, fmt.Errorf("transform: new_mod_convert: type %q: %v", conf.Type, errors.ErrInvalidOption)
	}

	tf := modConvert{
		conf: conf,
	}

	return &tf, nil
}

func (tf *modConvert) String() string {
	b, _ := gojson.Marshal(tf.conf)
	return string(b)
}

func (*modConvert) Close(context.Context) error {
	return nil
}

func (tf *modConvert) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	// Skip interrupt messages.
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	result := msg.GetObject(tf.conf.Object.Key)

	var value interface{}
	switch tf.conf.Type {
	case "bool":
		value = result.Bool()
	case "int":
		value = result.Int()
	case "float":
		value = result.Float()
	case "uint":
		value = result.Uint()
	case "string":
		value = result.String()
	}

	if err := msg.SetObject(tf.conf.Object.SetKey, value); err != nil {
		return nil, fmt.Errorf("transform: mod_convert: %v", err)
	}

	return []*message.Message{msg}, nil
}
