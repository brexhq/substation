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

type stringPatternFindAllConfig struct {
	Object iconfig.Object `json:"object"`

	// Pattern is the regular expression used to capture values.
	Pattern string `json:"pattern"`

	// Count is the number of matches to capture.
	Count int `json:"count"`

	re *regexp.Regexp
}

func (c *stringPatternFindAllConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *stringPatternFindAllConfig) Validate() error {
	if c.Object.Key == "" && c.Object.SetKey != "" {
		return fmt.Errorf("object_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.Object.Key != "" && c.Object.SetKey == "" {
		return fmt.Errorf("object_set_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.Pattern == "" {
		return fmt.Errorf("pattern: %v", errors.ErrMissingRequiredOption)
	}

	if c.Count == 0 {
		c.Count = -1
	}

	re, err := regexp.Compile(c.Pattern)
	if err != nil {
		return fmt.Errorf("pattern: %v", err)
	}

	c.re = re

	return nil
}

func newStringPatternFindAll(_ context.Context, cfg config.Config) (*stringPatternFindAll, error) {
	conf := stringPatternFindAllConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: new_str_capture_find_all: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: new_str_capture_find_all: %v", err)
	}

	tf := stringPatternFindAll{
		conf:     conf,
		isObject: conf.Object.Key != "" && conf.Object.SetKey != "",
	}

	return &tf, nil
}

type stringPatternFindAll struct {
	conf     stringPatternFindAllConfig
	isObject bool
}

func (tf *stringPatternFindAll) Transform(_ context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	if !tf.isObject {
		tmpMsg := message.New()

		subs := tf.conf.re.FindAllSubmatch(msg.Data(), tf.conf.Count)
		for _, s := range subs {
			m := strCaptureGetBytesMatch(s)
			if err := tmpMsg.SetValue("key.-1", m); err != nil {
				return nil, fmt.Errorf("transform: str_capture_find_all: %v", err)
			}
		}

		v := tmpMsg.GetValue("key")
		msg.SetData(v.Bytes())

		return []*message.Message{msg}, nil
	}

	v := msg.GetValue(tf.conf.Object.Key)
	subs := tf.conf.re.FindAllStringSubmatch(v.String(), tf.conf.Count)

	var matches []string
	for _, s := range subs {
		m := strCaptureGetStringMatch(s)
		matches = append(matches, m)
	}

	if err := msg.SetValue(tf.conf.Object.SetKey, matches); err != nil {
		return nil, fmt.Errorf("transform: str_capture_find_all: %v", err)
	}

	return []*message.Message{msg}, nil
}

func (tf *stringPatternFindAll) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
