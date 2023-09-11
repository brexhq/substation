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

type stringPatternNamedGroupConfig struct {
	Object iconfig.Object `json:"object"`

	// Expression is the regular expression used to capture values.
	Expression string `json:"expression"`

	re    *regexp.Regexp
	names []string
}

func (c *stringPatternNamedGroupConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *stringPatternNamedGroupConfig) Validate() error {
	if c.Object.Key == "" && c.Object.SetKey != "" {
		return fmt.Errorf("object_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.Object.Key != "" && c.Object.SetKey == "" {
		return fmt.Errorf("object_set_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.Expression == "" {
		return fmt.Errorf("expression: %v", errors.ErrMissingRequiredOption)
	}

	re, err := regexp.Compile(c.Expression)
	if err != nil {
		return fmt.Errorf("expression: %v", err)
	}

	c.re = re
	c.names = re.SubexpNames()

	return nil
}

func newStringPatternNamedGroup(_ context.Context, cfg config.Config) (*stringPatternNamedGroup, error) {
	conf := stringPatternNamedGroupConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: new_capture_named_group: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: new_capture_named_group: %v", err)
	}

	tf := stringPatternNamedGroup{
		conf:     conf,
		isObject: conf.Object.Key != "" && conf.Object.SetKey != "",
	}

	return &tf, nil
}

type stringPatternNamedGroup struct {
	conf     stringPatternNamedGroupConfig
	isObject bool
}

func (tf *stringPatternNamedGroup) Transform(_ context.Context, msg *message.Message) ([]*message.Message, error) {
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
				return nil, fmt.Errorf("transform: capture_named_group: %v", err)
			}
		}

		return []*message.Message{outMsg}, nil
	}

	v := msg.GetValue(tf.conf.Object.Key)
	matches := tf.conf.re.FindStringSubmatch(v.String())
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
			return nil, fmt.Errorf("transform: capture_named_group: %v", err)
		}
	}

	return []*message.Message{msg}, nil
}

func (tf *stringPatternNamedGroup) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
