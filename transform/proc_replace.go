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

type procReplaceConfig struct {
	// Key retrieves a value from an object for processing.
	//
	// This is optional for transforms that support processing non-object data.
	Key string `json:"key"`
	// SetKey inserts a processed value into an object.
	//
	// This is optional for transforms that support processing non-object data.
	SetKey string `json:"set_key"`
	// Old contains characters to replace in the data.
	Old string `json:"old"`
	// New contains characters that replace characters in Old.
	New string `json:"new"`
	// Counter determines the number of replacements to make.
	//
	// This is optional and defaults to -1 (replaces all matches).
	Count int `json:"count"`
}

type procReplace struct {
	conf     procReplaceConfig
	isObject bool
}

func newProcReplace(_ context.Context, cfg config.Config) (*procReplace, error) {
	conf := procReplaceConfig{}
	if err := _config.Decode(cfg.Settings, &conf); err != nil {
		return nil, err
	}

	// Validate required options.
	if (conf.Key != "" && conf.SetKey == "") ||
		(conf.Key == "" && conf.SetKey != "") {
		return nil, fmt.Errorf("transform: proc_replace: key %s set_key %s: %v", conf.Key, conf.SetKey, errInvalidDataPattern)
	}

	if conf.Old == "" {
		return nil, fmt.Errorf("transform: proc_replace: old: %v", errors.ErrMissingRequiredOption)
	}

	if conf.Count == 0 {
		conf.Count = -1
	}

	proc := procReplace{
		conf:     conf,
		isObject: conf.Key != "" && conf.SetKey != "",
	}

	return &proc, nil
}

func (t *procReplace) String() string {
	b, _ := gojson.Marshal(t.conf)
	return string(b)
}

func (*procReplace) Close(context.Context) error {
	return nil
}

func (t *procReplace) Transform(ctx context.Context, messages ...*mess.Message) ([]*mess.Message, error) {
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
			value := strings.Replace(
				result,
				t.conf.Old,
				t.conf.New,
				t.conf.Count,
			)

			if err := message.Set(t.conf.SetKey, value); err != nil {
				return nil, fmt.Errorf("transform: proc_replace: %v", err)
			}

			output = append(output, message)

		case false:
			value := bytes.Replace(
				message.Data(),
				[]byte(t.conf.Old),
				[]byte(t.conf.New),
				t.conf.Count,
			)

			msg, err := mess.New(
				mess.SetData(value),
				mess.SetMetadata(message.Metadata()),
			)
			if err != nil {
				return nil, fmt.Errorf("transform: proc_replace: %v", err)
			}

			output = append(output, msg)
		}
	}

	return output, nil
}
