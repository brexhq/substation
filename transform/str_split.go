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

type strSplitConfig struct {
	Object iconfig.Object `json:"object"`

	// Separator splits the string into elements of the array.
	Separator string `json:"separator"`
}

func (c *strSplitConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *strSplitConfig) Validate() error {
	if c.Object.Key == "" && c.Object.SetKey != "" {
		return fmt.Errorf("object_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.Object.Key != "" && c.Object.SetKey == "" {
		return fmt.Errorf("object_set_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.Separator == "" {
		return fmt.Errorf("separator: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

type strSplit struct {
	conf     strSplitConfig
	isObject bool

	separator []byte
}

func newStrSplit(_ context.Context, cfg config.Config) (*strSplit, error) {
	conf := strSplitConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("strings: new_split: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("strings: new_split: %v", err)
	}

	tf := strSplit{
		conf:      conf,
		isObject:  conf.Object.Key != "" && conf.Object.SetKey != "",
		separator: []byte(conf.Separator),
	}

	return &tf, nil
}

func (tf *strSplit) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	// Skip interrupt messages.
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	if !tf.isObject {
		tmpMsg := message.New()

		b := bytes.Split(msg.Data(), tf.separator)
		for _, v := range b {
			if err := tmpMsg.SetValue("key.-1", v); err != nil {
				return nil, fmt.Errorf("strings: split: %v", err)
			}
		}

		value := tmpMsg.GetValue("key")
		msg.SetData(value.Bytes())

		return []*message.Message{msg}, nil
	}

	value := msg.GetValue(tf.conf.Object.Key)
	str := strings.Split(value.String(), tf.conf.Separator)

	if err := msg.SetValue(tf.conf.Object.SetKey, str); err != nil {
		return nil, fmt.Errorf("strings: split: %v", err)
	}

	return []*message.Message{msg}, nil
}

func (tf *strSplit) String() string {
	b, _ := gojson.Marshal(tf.conf)
	return string(b)
}

func (*strSplit) Close(context.Context) error {
	return nil
}
