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

type stringSplitConfig struct {
	// Separator splits the string into elements of the array.
	Separator string `json:"separator"`

	Object iconfig.Object `json:"object"`
}

func (c *stringSplitConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *stringSplitConfig) Validate() error {
	if c.Object.SourceKey == "" && c.Object.TargetKey != "" {
		return fmt.Errorf("object_source_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.Object.SourceKey != "" && c.Object.TargetKey == "" {
		return fmt.Errorf("object_target_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.Separator == "" {
		return fmt.Errorf("separator: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

type stringSplit struct {
	conf     stringSplitConfig
	isObject bool

	separator []byte
}

func newStringSplit(_ context.Context, cfg config.Config) (*stringSplit, error) {
	conf := stringSplitConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: string_split: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: string_split: %v", err)
	}

	tf := stringSplit{
		conf:      conf,
		isObject:  conf.Object.SourceKey != "" && conf.Object.TargetKey != "",
		separator: []byte(conf.Separator),
	}

	return &tf, nil
}

func (tf *stringSplit) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	if !tf.isObject {
		tmpMsg := message.New()

		b := bytes.Split(msg.Data(), tf.separator)
		for _, v := range b {
			if err := tmpMsg.SetValue("key.-1", v); err != nil {
				return nil, fmt.Errorf("transform: string_split: %v", err)
			}
		}

		value := tmpMsg.GetValue("key")
		msg.SetData(value.Bytes())

		return []*message.Message{msg}, nil
	}

	value := msg.GetValue(tf.conf.Object.SourceKey)
	if !value.Exists() {
		return []*message.Message{msg}, nil
	}

	str := strings.Split(value.String(), tf.conf.Separator)

	if err := msg.SetValue(tf.conf.Object.TargetKey, str); err != nil {
		return nil, fmt.Errorf("transform: string_split: %v", err)
	}

	return []*message.Message{msg}, nil
}

func (tf *stringSplit) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
