package transform

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/brexhq/substation/config"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/message"
)

type stringMatchFindConfig struct {
	Object iconfig.Object `json:"object"`

	// Pattern is the regular expression used to capture values.
	Pattern string `json:"pattern"`

	re *regexp.Regexp
}

func (c *stringMatchFindConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *stringMatchFindConfig) Validate() error {
	if c.Object.Key == "" && c.Object.SetKey != "" {
		return fmt.Errorf("object_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.Object.Key != "" && c.Object.SetKey == "" {
		return fmt.Errorf("object_set_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.Pattern == "" {
		return fmt.Errorf("pattern: %v", errors.ErrMissingRequiredOption)
	}

	re, err := regexp.Compile(c.Pattern)
	if err != nil {
		return fmt.Errorf("pattern: %v", err)
	}

	c.re = re

	return nil
}

func newStringMatchFind(_ context.Context, cfg config.Config) (*stringMatchFind, error) {
	conf := stringMatchFindConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: string_match_find: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: string_match_find: %v", err)
	}

	tf := stringMatchFind{
		conf:     conf,
		isObject: conf.Object.Key != "" && conf.Object.SetKey != "",
	}

	return &tf, nil
}

type stringMatchFind struct {
	conf     stringMatchFindConfig
	isObject bool
}

func (tf *stringMatchFind) Transform(_ context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	if !tf.isObject {
		matches := tf.conf.re.FindSubmatch(msg.Data())
		msg.SetData(strCaptureGetBytesMatch(matches))

		return []*message.Message{msg}, nil
	}

	value := msg.GetValue(tf.conf.Object.Key)
	if !value.Exists() {
		return []*message.Message{msg}, nil
	}

	matches := tf.conf.re.FindStringSubmatch(value.String())

	if err := msg.SetValue(tf.conf.Object.SetKey, strCaptureGetStringMatch(matches)); err != nil {
		return nil, fmt.Errorf("transform: string_match_find: %v", err)
	}

	return []*message.Message{msg}, nil
}

func (tf *stringMatchFind) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
