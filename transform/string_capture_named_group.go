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

type stringCaptureNamedGroupConfig struct {
	Object iconfig.Object `json:"object"`

	// Pattern is the regular expression used to capture values.
	Pattern string `json:"pattern"`

	re    *regexp.Regexp
	names []string
}

func (c *stringCaptureNamedGroupConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *stringCaptureNamedGroupConfig) Validate() error {
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
	c.names = re.SubexpNames()

	return nil
}

func newStringCaptureNamedGroup(_ context.Context, cfg config.Config) (*stringCaptureNamedGroup, error) {
	conf := stringCaptureNamedGroupConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: string_match_named_group: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: string_match_named_group: %v", err)
	}

	tf := stringCaptureNamedGroup{
		conf:     conf,
		isObject: conf.Object.Key != "" && conf.Object.SetKey != "",
	}

	return &tf, nil
}

type stringCaptureNamedGroup struct {
	conf     stringCaptureNamedGroupConfig
	isObject bool
}

func (tf *stringCaptureNamedGroup) Transform(_ context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	if !tf.isObject {
		outMsg := message.New().SetMetadata(msg.Metadata())

		matches := tf.conf.re.FindSubmatch(msg.Data())
		for i, m := range matches {
			if i == 0 {
				continue
			}

			if err := outMsg.SetValue(tf.conf.names[i], m); err != nil {
				return nil, fmt.Errorf("transform: string_match_named_group: %v", err)
			}
		}

		return []*message.Message{outMsg}, nil
	}

	value := msg.GetValue(tf.conf.Object.Key)
	if !value.Exists() {
		return []*message.Message{msg}, nil
	}

	matches := tf.conf.re.FindStringSubmatch(value.String())
	for i, match := range matches {
		if i == 0 {
			continue
		}

		// If the same key is used multiple times, then this will correctly
		// set multiple named groups into that key.
		//
		// If set_key is "a" and the first group returns {"b":"c"}, then
		// the output is {"a":{"b":"c"}}. If the second group returns
		// {"d":"e"} then the output is {"a":{"b":"c","d":"e"}}.
		setKey := tf.conf.Object.SetKey + "." + tf.conf.names[i]
		if err := msg.SetValue(setKey, match); err != nil {
			return nil, fmt.Errorf("transform: string_match_named_group: %v", err)
		}
	}

	return []*message.Message{msg}, nil
}

func (tf *stringCaptureNamedGroup) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
