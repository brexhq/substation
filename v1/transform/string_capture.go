package transform

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/brexhq/substation/config"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/message"
)

type stringCaptureConfig struct {
	// Pattern is the regular expression used to capture values.
	Pattern string `json:"pattern"`
	re      *regexp.Regexp

	// Count is the number of captures to make.
	//
	// This is optional and defaults to 0, which means that a single
	// capture is made. If a named capture group is used, then this
	// is ignored.
	Count int `json:"count"`

	ID     string         `json:"id"`
	Object iconfig.Object `json:"object"`
}

func (c *stringCaptureConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *stringCaptureConfig) Validate() error {
	if c.Object.SourceKey == "" && c.Object.TargetKey != "" {
		return fmt.Errorf("object_source_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.Object.SourceKey != "" && c.Object.TargetKey == "" {
		return fmt.Errorf("object_target_key: %v", errors.ErrMissingRequiredOption)
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
		return nil, fmt.Errorf("transform string_capture: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "string_capture"
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	tf := stringCapture{
		conf:     conf,
		isObject: conf.Object.SourceKey != "" && conf.Object.TargetKey != "",
		// Check if the regular expression contains at least one named capture group.
		containsCaptureGroup: strings.Contains(conf.Pattern, "(?P<"),
	}

	return &tf, nil
}

type stringCapture struct {
	conf                 stringCaptureConfig
	isObject             bool
	containsCaptureGroup bool
}

func (tf *stringCapture) Transform(_ context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	if !tf.isObject {
		switch {
		case tf.containsCaptureGroup:
			outMsg := message.New().SetMetadata(msg.Metadata())

			matches := tf.conf.re.FindSubmatch(msg.Data())
			for i, m := range matches {
				if i == 0 {
					continue
				}

				if err := outMsg.SetValue(tf.conf.re.SubexpNames()[i], m); err != nil {
					return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
				}
			}

			return []*message.Message{outMsg}, nil

		case tf.conf.Count == 0:
			matches := tf.conf.re.FindSubmatch(msg.Data())
			msg.SetData(strCaptureGetBytesMatch(matches))

			return []*message.Message{msg}, nil

		default:
			tmpMsg := message.New()
			subs := tf.conf.re.FindAllSubmatch(msg.Data(), tf.conf.Count)

			for _, s := range subs {
				m := strCaptureGetBytesMatch(s)
				if err := tmpMsg.SetValue("key.-1", m); err != nil {
					return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
				}
			}

			v := tmpMsg.GetValue("key")
			msg.SetData(v.Bytes())

			return []*message.Message{msg}, nil
		}
	}

	value := msg.GetValue(tf.conf.Object.SourceKey)
	if !value.Exists() {
		return []*message.Message{msg}, nil
	}

	switch {
	case tf.containsCaptureGroup:
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
			setKey := tf.conf.Object.TargetKey + "." + tf.conf.re.SubexpNames()[i]
			if err := msg.SetValue(setKey, match); err != nil {
				return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
			}
		}

		return []*message.Message{msg}, nil

	case tf.conf.Count == 0:
		matches := tf.conf.re.FindStringSubmatch(value.String())
		if err := msg.SetValue(tf.conf.Object.TargetKey, strCaptureGetStringMatch(matches)); err != nil {
			return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
		}

		return []*message.Message{msg}, nil

	default:
		var matches []string
		subs := tf.conf.re.FindAllStringSubmatch(value.String(), tf.conf.Count)

		for _, s := range subs {
			m := strCaptureGetStringMatch(s)
			matches = append(matches, m)
		}

		if err := msg.SetValue(tf.conf.Object.TargetKey, matches); err != nil {
			return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
		}

		return []*message.Message{msg}, nil
	}
}

func (tf *stringCapture) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
