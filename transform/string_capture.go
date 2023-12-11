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

type stringCaptureConfig struct {
	Object iconfig.Object `json:"object"`

	// Pattern is the regular expression used to capture values.
	Pattern string `json:"pattern"`

	// Count is the number of matches to capture.
	//
	// This is optional and defaults to 0, which means that a single
	// capture is made.
	Count int `json:"count"`

	re *regexp.Regexp
}

func (c *stringCaptureConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *stringCaptureConfig) Validate() error {
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

func newStringCapture(_ context.Context, cfg config.Config) (*stringCapture, error) {
	conf := stringCaptureConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: string_capture: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: string_capture: %v", err)
	}

	tf := stringCapture{
		conf:     conf,
		isObject: conf.Object.Key != "" && conf.Object.SetKey != "",
	}

	return &tf, nil
}

type stringCapture struct {
	conf     stringCaptureConfig
	isObject bool
}

func (tf *stringCapture) Transform(_ context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	if !tf.isObject {
		tmpMsg := message.New()

		if tf.conf.Count == 0 {
			matches := tf.conf.re.FindSubmatch(msg.Data())
			msg.SetData(strCaptureGetBytesMatch(matches))

			return []*message.Message{msg}, nil
		}

		subs := tf.conf.re.FindAllSubmatch(msg.Data(), tf.conf.Count)
		for _, s := range subs {
			m := strCaptureGetBytesMatch(s)
			if err := tmpMsg.SetValue("key.-1", m); err != nil {
				return nil, fmt.Errorf("transform: string_capture: %v", err)
			}
		}

		v := tmpMsg.GetValue("key")
		msg.SetData(v.Bytes())

		return []*message.Message{msg}, nil
	}

	value := msg.GetValue(tf.conf.Object.Key)
	if !value.Exists() {
		return []*message.Message{msg}, nil
	}

	if tf.conf.Count == 0 {
		matches := tf.conf.re.FindStringSubmatch(value.String())
		if err := msg.SetValue(tf.conf.Object.SetKey, strCaptureGetStringMatch(matches)); err != nil {
			return nil, fmt.Errorf("transform: string_capture: %v", err)
		}

		return []*message.Message{msg}, nil
	}

	subs := tf.conf.re.FindAllStringSubmatch(value.String(), tf.conf.Count)

	var matches []string
	for _, s := range subs {
		m := strCaptureGetStringMatch(s)
		matches = append(matches, m)
	}

	if err := msg.SetValue(tf.conf.Object.SetKey, matches); err != nil {
		return nil, fmt.Errorf("transform: string_capture: %v", err)
	}

	return []*message.Message{msg}, nil
}

func (tf *stringCapture) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
