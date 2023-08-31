package transform

import (
	"context"
	gojson "encoding/json"
	"fmt"
	"regexp"

	"golang.org/x/exp/slices"

	"github.com/brexhq/substation/config"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/message"
)

type modCaptureConfig struct {
	Object configObject `json:"object"`

	// Expression is the regular expression used to capture values.
	Expression string `json:"expression"`
	// Type determines which regular expression function is applied using
	// the Expression.
	//
	// Must be one of:
	//	- find: applies the Find(String)?Submatch function
	//	- find_all: applies the FindAll(String)?Submatch function (see count)
	//	- named_group: applies the Find(String)?Submatch function and stores
	//	values as objects using subexpressions
	Type string `json:"type"`
	// Count manages the number of repeated capture groups.
	//
	// This is optional and defaults to match all capture groups.
	Count int `json:"count"`
}

type modCapture struct {
	conf  modCaptureConfig
	isObj bool

	re *regexp.Regexp
}

func newModCapture(_ context.Context, cfg config.Config) (*modCapture, error) {
	conf := modCaptureConfig{}
	if err := iconfig.Decode(cfg.Settings, &conf); err != nil {
		return nil, fmt.Errorf("transform: new_mod_capture: %v", err)
	}

	// Validate required options.
	if conf.Object.Key == "" && conf.Object.SetKey != "" {
		return nil, fmt.Errorf("transform: new_mod_capture: object_key: %v", errors.ErrMissingRequiredOption)
	}

	if conf.Object.Key != "" && conf.Object.SetKey == "" {
		return nil, fmt.Errorf("transform: new_mod_capture: object_set_key: %v", errors.ErrMissingRequiredOption)
	}

	if conf.Type == "" {
		return nil, fmt.Errorf("transform: new_mod_capture: type: %v", errors.ErrMissingRequiredOption)
	}

	if !slices.Contains(
		[]string{
			"find",
			"find_all",
			"named_group",
		},
		conf.Type) {
		return nil, fmt.Errorf("transform: new_mod_capture: type %q: %v", conf.Type, errors.ErrInvalidOption)
	}

	if conf.Expression == "" {
		return nil, fmt.Errorf("transform: new_mod_capture: expression: %v", errors.ErrMissingRequiredOption)
	}

	if conf.Count == 0 {
		conf.Count = -1
	}

	re, err := regexp.Compile(conf.Expression)
	if err != nil {
		return nil, fmt.Errorf("transform: new_mod_capture: expression: %v", err)
	}

	tf := modCapture{
		conf:  conf,
		isObj: conf.Object.Key != "" && conf.Object.SetKey != "",
		re:    re,
	}

	return &tf, nil
}

func (tf *modCapture) String() string {
	b, _ := gojson.Marshal(tf.conf)
	return string(b)
}

func (*modCapture) Close(context.Context) error {
	return nil
}

func (tf *modCapture) Transform(_ context.Context, msg *message.Message) ([]*message.Message, error) {
	// Skip interrupt messages.
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	if !tf.isObj {
		switch tf.conf.Type {
		case "find":
			match := tf.re.FindSubmatch(msg.Data())
			outMsg := message.New().SetData(match[1]).SetMetadata(msg.Metadata())

			return []*message.Message{outMsg}, nil
		case "named_group":
			names := tf.re.SubexpNames()
			matches := tf.re.FindSubmatch(msg.Data())

			outMsg := message.New().SetMetadata(msg.Metadata())
			for i, m := range matches {
				if i == 0 {
					continue
				}

				if err := outMsg.SetObject(names[i], m); err != nil {
					return nil, fmt.Errorf("transform: mod_capture: %v", err)
				}
			}

			return []*message.Message{outMsg}, nil
		}
	}

	result := msg.GetObject(tf.conf.Object.Key).String()

	switch tf.conf.Type {
	case "find":
		match := tf.re.FindStringSubmatch(result)
		if err := msg.SetObject(tf.conf.Object.SetKey, tf.getStringMatch(match)); err != nil {
			return nil, fmt.Errorf("transform: mod_capture: %v", err)
		}
	case "find_all":
		var matches []interface{}

		subs := tf.re.FindAllStringSubmatch(result, tf.conf.Count)
		for _, s := range subs {
			m := tf.getStringMatch(s)
			matches = append(matches, m)
		}

		if err := msg.SetObject(tf.conf.Object.SetKey, matches); err != nil {
			return nil, fmt.Errorf("transform: mod_capture: %v", err)
		}
	case "named_group":
		names := tf.re.SubexpNames()
		matches := tf.re.FindStringSubmatch(result)
		for i, m := range matches {
			if i == 0 {
				continue
			}

			// If the same key is used multiple times, then this will correctly
			// set multiple named groups into that key.
			//
			// If set_key is "a" and the first group returns {"b":"c"}, then
			// the output is {"a":{"b":"c"}}. If the second group returns
			// {"d":"e"} then the output is {"a":{"b":"c","d":"e"}}.
			setKey := tf.conf.Object.SetKey + "." + names[i]
			if err := msg.SetObject(setKey, m); err != nil {
				return nil, fmt.Errorf("transform: mod_capture: %v", err)
			}
		}
	}

	return []*message.Message{msg}, nil
}

func (tf *modCapture) getStringMatch(match []string) string {
	if len(match) > 1 {
		return match[len(match)-1]
	}

	return ""
}
