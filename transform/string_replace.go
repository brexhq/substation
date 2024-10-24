package transform

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"

	iconfig "github.com/brexhq/substation/v2/internal/config"
)

type stringReplaceConfig struct {
	// Pattern is the regular expression used to identify values to replace.
	Pattern string `json:"pattern"`
	re      *regexp.Regexp
	// Replacement is the string to replace the matched values with.
	Replacement string `json:"replacement"`

	ID     string         `json:"id"`
	Object iconfig.Object `json:"object"`
}

func (c *stringReplaceConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *stringReplaceConfig) Validate() error {
	if c.Object.SourceKey == "" && c.Object.TargetKey != "" {
		return fmt.Errorf("object_source_key: %v", iconfig.ErrMissingRequiredOption)
	}

	if c.Object.SourceKey != "" && c.Object.TargetKey == "" {
		return fmt.Errorf("object_target_key: %v", iconfig.ErrMissingRequiredOption)
	}

	if c.Pattern == "" {
		return fmt.Errorf("old: %v", iconfig.ErrMissingRequiredOption)
	}

	re, err := regexp.Compile(c.Pattern)
	if err != nil {
		return fmt.Errorf("pattern: %v", err)
	}

	c.re = re

	return nil
}

func newStringReplace(_ context.Context, cfg config.Config) (*stringReplace, error) {
	conf := stringReplaceConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform string_replace: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "string_replace"
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	tf := stringReplace{
		conf:     conf,
		isObject: conf.Object.SourceKey != "" && conf.Object.TargetKey != "",
		r:        []byte(conf.Replacement),
	}

	return &tf, nil
}

type stringReplace struct {
	conf     stringReplaceConfig
	isObject bool

	r []byte
}

func (tf *stringReplace) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.HasFlag(message.IsControl) {
		return []*message.Message{msg}, nil
	}

	if !tf.isObject {
		b := tf.conf.re.ReplaceAll(msg.Data(), tf.r)
		msg.SetData(b)

		return []*message.Message{msg}, nil
	}

	value := msg.GetValue(tf.conf.Object.SourceKey)
	if skipMessage(msg, value) {
		return []*message.Message{msg}, nil
	}

	s := tf.conf.re.ReplaceAllString(value.String(), string(tf.r))
	if err := msg.SetValue(tf.conf.Object.TargetKey, s); err != nil {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
	}

	return []*message.Message{msg}, nil
}

func (tf *stringReplace) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
