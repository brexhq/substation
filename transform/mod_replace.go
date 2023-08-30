package transform

import (
	"bytes"
	"context"
	gojson "encoding/json"
	"fmt"
	"strings"

	"github.com/brexhq/substation/config"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/message"
)

type modReplaceConfig struct {
	Object configObject `json:"object"`

	// Old contains characters to replace in the data.
	Old string `json:"old"`
	// New contains characters that replace characters in Old.
	New string `json:"new"`
	// Counter determines the number of replacements to make.
	//
	// This is optional and defaults to -1 (replaces all matches).
	Count int `json:"count"`
}

type modReplace struct {
	conf  modReplaceConfig
	isObj bool

	old []byte
	new []byte
}

func newModReplace(_ context.Context, cfg config.Config) (*modReplace, error) {
	conf := modReplaceConfig{}
	if err := iconfig.Decode(cfg.Settings, &conf); err != nil {
		return nil, fmt.Errorf("transform: new_mod_replace: %v", err)
	}

	// Validate required options.
	if conf.Object.Key == "" && conf.Object.SetKey != "" {
		return nil, fmt.Errorf("transform: new_mod_replace: object_key: %v", errors.ErrMissingRequiredOption)
	}

	if conf.Object.Key != "" && conf.Object.SetKey == "" {
		return nil, fmt.Errorf("transform: new_mod_replace: object_set_key: %v", errors.ErrMissingRequiredOption)
	}

	if conf.Old == "" {
		return nil, fmt.Errorf("transform: new_mod_replace: old: %v", errors.ErrMissingRequiredOption)
	}

	if conf.Count == 0 {
		conf.Count = -1
	}

	tf := modReplace{
		conf:  conf,
		isObj: conf.Object.Key != "" && conf.Object.SetKey != "",
		old:   []byte(conf.Old),
		new:   []byte(conf.New),
	}

	return &tf, nil
}

func (tf *modReplace) String() string {
	b, _ := gojson.Marshal(tf.conf)
	return string(b)
}

func (*modReplace) Close(context.Context) error {
	return nil
}

func (tf *modReplace) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	// Skip interrupt messages.
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	if !tf.isObj {
		value := bytes.Replace(
			msg.Data(),
			tf.old,
			tf.new,
			tf.conf.Count,
		)

		finMsg := message.New().SetData(value).SetMetadata(msg.Metadata())
		return []*message.Message{finMsg}, nil
	}

	result := msg.GetObject(tf.conf.Object.Key).String()
	value := strings.Replace(
		result,
		tf.conf.Old,
		tf.conf.New,
		tf.conf.Count,
	)

	if err := msg.SetObject(tf.conf.Object.SetKey, value); err != nil {
		return nil, fmt.Errorf("transform: mod_replace: %v", err)
	}

	return []*message.Message{msg}, nil
}
