package transform

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/brexhq/substation/config"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/message"
)

type strReplaceConfig struct {
	Object iconfig.Object `json:"object"`

	// Old contains characters to replace in the data.
	Old string `json:"old"`
	// New contains characters that replace characters in Old.
	New string `json:"new"`
	// Counter determines the number of replacements to make.
	//
	// This is optional and defaults to -1 (replaces all matches).
	Count int `json:"count"`
}

func (c *strReplaceConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *strReplaceConfig) Validate() error {
	if c.Object.Key == "" && c.Object.SetKey != "" {
		return fmt.Errorf("object_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.Object.Key != "" && c.Object.SetKey == "" {
		return fmt.Errorf("object_set_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.Old == "" {
		return fmt.Errorf("transform: new_mod_replace: old: %v", errors.ErrMissingRequiredOption)
	}

	if c.Count == 0 {
		c.Count = -1
	}

	return nil
}

type strReplace struct {
	conf     strReplaceConfig
	isObject bool

	old []byte
	new []byte
}

func newStrReplace(_ context.Context, cfg config.Config) (*strReplace, error) {
	conf := strReplaceConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: new_str_replace: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: new_str_replace: %v", err)
	}

	tf := strReplace{
		conf:     conf,
		isObject: conf.Object.Key != "" && conf.Object.SetKey != "",
		old:      []byte(conf.Old),
		new:      []byte(conf.New),
	}

	return &tf, nil
}

func (tf *strReplace) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	// Skip interrupt messages.
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	if !tf.isObject {
		b := bytes.Replace(msg.Data(), tf.old, tf.new, tf.conf.Count)
		msg.SetData(b)

		return []*message.Message{msg}, nil
	}

	value := msg.GetValue(tf.conf.Object.Key)
	if !value.Exists() {
		return []*message.Message{msg}, nil
	}

	s := strings.Replace(value.String(), tf.conf.Old, tf.conf.New, tf.conf.Count)
	if err := msg.SetValue(tf.conf.Object.SetKey, s); err != nil {
		return nil, fmt.Errorf("transform: str_replace: %v", err)
	}

	return []*message.Message{msg}, nil
}

func (tf *strReplace) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}

func (*strReplace) Close(context.Context) error {
	return nil
}
