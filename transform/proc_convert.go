package transform

import (
	"context"
	gojson "encoding/json"
	"fmt"

	"golang.org/x/exp/slices"

	"github.com/brexhq/substation/config"
	_config "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	mess "github.com/brexhq/substation/message"
)

type procConvertConfig struct {
	// Key retrieves a value from an object for processing.
	//
	// This is optional for transforms that support processing non-object data.
	Key string `json:"key"`
	// SetKey inserts a processed value into an object.
	//
	// This is optional for transforms that support processing non-object data.
	SetKey string `json:"set_key"`
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

type procConvert struct {
	conf procConvertConfig
}

func newProcConvert(_ context.Context, cfg config.Config) (*procConvert, error) {
	conf := procConvertConfig{}
	if err := _config.Decode(cfg.Settings, &conf); err != nil {
		return nil, err
	}

	// Validate required options.
	if conf.Key == "" || conf.SetKey == "" {
		return nil, fmt.Errorf("transform: proc_convert: key %s set_key %s: %v", conf.Key, conf.SetKey, errInvalidDataPattern)
	}

	if conf.Type == "" {
		return nil, fmt.Errorf("transform: proc_convert: type: %v", errors.ErrMissingRequiredOption)
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
		return nil, fmt.Errorf("transform: proc_convert: type %q: %v", conf.Type, errors.ErrInvalidOption)
	}

	p := procConvert{
		conf: conf,
	}

	return &p, nil
}

func (proc *procConvert) String() string {
	b, _ := gojson.Marshal(proc.conf)
	return string(b)
}

func (*procConvert) Close(context.Context) error {
	return nil
}

func (proc *procConvert) Transform(ctx context.Context, message *mess.Message) ([]*mess.Message, error) {
	// Skip control messages.
	if message.IsControl() {
		return []*mess.Message{message}, nil
	}

	result := message.Get(proc.conf.Key)

	var value interface{}
	switch proc.conf.Type {
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

	if err := message.Set(proc.conf.SetKey, value); err != nil {
		return nil, fmt.Errorf("transform: proc_convert: %v", err)
	}

	return []*mess.Message{message}, nil
}
