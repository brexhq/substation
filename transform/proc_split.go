package transform

import (
	"bytes"
	"context"
	gojson "encoding/json"
	"fmt"
	"strings"

	"github.com/brexhq/substation/config"
	_config "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	mess "github.com/brexhq/substation/message"
)

type procSplitConfig struct {
	// Key retrieves a value from an object for processing.
	//
	// This is optional for transforms that support processing non-object data.
	Key string `json:"key"`
	// SetKey inserts a processed value into an object.
	//
	// This is optional for transforms that support processing non-object data.
	SetKey string `json:"set_key"`
	// Separator is the string that splits data.
	Separator string `json:"separator"`
}

type procSplit struct {
	conf     procSplitConfig
	isObject bool
}

func newProcSplit(_ context.Context, cfg config.Config) (*procSplit, error) {
	conf := procSplitConfig{}
	if err := _config.Decode(cfg.Settings, &conf); err != nil {
		return nil, err
	}

	// Validate required options.
	if (conf.Key != "" && conf.SetKey == "") ||
		(conf.Key == "" && conf.SetKey != "") {
		return nil, fmt.Errorf("transform: proc_split: key %s set_key %s: %v", conf.Key, conf.SetKey, errInvalidDataPattern)
	}

	if conf.Separator == "" {
		return nil, fmt.Errorf("transform: proc_split: separator: %v", errors.ErrMissingRequiredOption)
	}

	proc := procSplit{
		conf:     conf,
		isObject: conf.Key != "" && conf.SetKey != "",
	}

	return &proc, nil
}

func (proc *procSplit) String() string {
	b, _ := gojson.Marshal(proc.conf)
	return string(b)
}

func (*procSplit) Close(context.Context) error {
	return nil
}

func (proc *procSplit) Transform(ctx context.Context, message *mess.Message) ([]*mess.Message, error) {
	// Skip control messages.
	if message.IsControl() {
		return []*mess.Message{message}, nil
	}

	if !proc.isObject {
		var output []*mess.Message

		for _, x := range bytes.Split(message.Data(), []byte(proc.conf.Separator)) {
			msg, err := mess.New(
				mess.SetData(x),
				mess.SetMetadata(message.Metadata()),
			)
			if err != nil {
				return nil, fmt.Errorf("transform: proc_split: %v", err)
			}

			output = append(output, msg)
		}

		return output, nil
	}

	res := message.Get(proc.conf.Key).String()
	v := strings.Split(res, proc.conf.Separator)

	if err := message.Set(proc.conf.SetKey, v); err != nil {
		return nil, fmt.Errorf("transform: proc_split: %v", err)
	}

	return []*mess.Message{message}, nil
}
