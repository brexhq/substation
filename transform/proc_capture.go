package transform

import (
	"context"
	gojson "encoding/json"
	"fmt"
	"regexp"

	"golang.org/x/exp/slices"

	"github.com/brexhq/substation/config"
	_config "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	mess "github.com/brexhq/substation/message"
)

type procCaptureConfig struct {
	// Key retrieves a value from an object for processing.
	//
	// This is optional for transforms that support processing non-object data.
	Key string `json:"key"`
	// SetKey inserts a processed value into an object.
	//
	// This is optional for transforms that support processing non-object data.
	SetKey string `json:"set_key"`
	// Expression is the regular expression used to capture values.
	Expression string `json:"expression"`
	// Type determines which regular expression function is applied using
	// the Expression.
	//
	// Must be one of:
	//
	// - find: applies the Find(String)?Submatch function
	//
	// - find_all: applies the FindAll(String)?Submatch function (see count)
	//
	// - named_group: applies the Find(String)?Submatch function and stores
	// values as objects using subexpressions
	Type string `json:"type"`
	// Count manages the number of repeated capture groups.
	//
	// This is optional and defaults to match all capture groups.
	Count int `json:"count"`
}

type procCapture struct {
	conf     procCaptureConfig
	isObject bool

	re *regexp.Regexp
}

func newProcCapture(_ context.Context, cfg config.Config) (*procCapture, error) {
	conf := procCaptureConfig{}
	if err := _config.Decode(cfg.Settings, &conf); err != nil {
		return nil, err
	}

	// Validate required options.
	if (conf.Key == "" && conf.SetKey != "") || (conf.Key != "" && conf.SetKey == "") {
		return nil, fmt.Errorf("transform: proc_capture: key %s set_key %s: %v", conf.Key, conf.SetKey, errInvalidDataPattern)
	}

	if conf.Type == "" {
		return nil, fmt.Errorf("transform: proc_capture: type: %v", errors.ErrMissingRequiredOption)
	}

	if !slices.Contains(
		[]string{
			"find",
			"find_all",
			"named_group",
		},
		conf.Type) {
		return nil, fmt.Errorf("transform: proc_capture: type %q: %v", conf.Type, errors.ErrInvalidOption)
	}

	if conf.Expression == "" {
		return nil, fmt.Errorf("transform: proc_capture: option \"expression\": %v", errors.ErrMissingRequiredOption)
	}

	if conf.Count == 0 {
		conf.Count = -1
	}

	re, err := regexp.Compile(conf.Expression)
	if err != nil {
		return nil, fmt.Errorf("transform: proc_capture: %v", err)
	}

	proc := procCapture{
		conf:     conf,
		isObject: conf.Key != "" && conf.SetKey != "",
		re:       re,
	}

	return &proc, nil
}

func (t *procCapture) String() string {
	b, _ := gojson.Marshal(t.conf)
	return string(b)
}

func (*procCapture) Close(context.Context) error {
	return nil
}

//nolint: gocognit // Ignore cognitive complexity.
func (t *procCapture) Transform(_ context.Context, messages ...*mess.Message) ([]*mess.Message, error) {
	var output []*mess.Message

	for _, message := range messages {
		// Skip control messages.
		if message.IsControl() {
			output = append(output, message)
			continue
		}

		switch t.isObject {
		case true:
			result := message.Get(t.conf.Key).String()

			switch t.conf.Type {
			case "find":
				match := t.re.FindStringSubmatch(result)
				if err := message.Set(t.conf.SetKey, t.getStringMatch(match)); err != nil {
					return nil, fmt.Errorf("transform: proc_capture: %v", err)
				}
			case "find_all":
				var matches []interface{}

				subs := t.re.FindAllStringSubmatch(result, t.conf.Count)
				for _, s := range subs {
					m := t.getStringMatch(s)
					matches = append(matches, m)
				}

				if err := message.Set(t.conf.SetKey, matches); err != nil {
					return nil, fmt.Errorf("transform: proc_capture: %v", err)
				}
			case "named_group":
				names := t.re.SubexpNames()
				matches := t.re.FindStringSubmatch(result)
				for i, m := range matches {
					if i == 0 {
						continue
					}

					// If the same key is used multiple times, then this will correctly
					// set multiple named groups into that key.
					//
					// If set_key is "foo" and the first group returns {"bar":"baz"}, then
					// the output is {"foo":{"bar":"baz"}}. If the second group returns
					// {"qux":"quux"} then the output is {"foo":{"bar":"baz","qux":"quux"}}.
					setKey := t.conf.SetKey + "." + names[i]
					if err := message.Set(setKey, m); err != nil {
						return nil, fmt.Errorf("transform: proc_capture: %v", err)
					}
				}
			}

			output = append(output, message)
		case false:
			switch t.conf.Type {
			case "find":
				match := t.re.FindSubmatch(message.Data())
				msg, err := mess.New(
					mess.SetData(match[1]),
				)
				if err != nil {
					return nil, fmt.Errorf("transform: proc_capture: %v", err)
				}

				output = append(output, msg)
			case "named_group":
				msg, err := mess.New()
				if err != nil {
					return nil, fmt.Errorf("transform: proc_capture: %v", err)
				}

				names := t.re.SubexpNames()
				matches := t.re.FindSubmatch(message.Data())
				for i, m := range matches {
					if i == 0 {
						continue
					}

					if err := msg.Set(names[i], m); err != nil {
						return nil, fmt.Errorf("transform: proc_capture: %v", err)
					}
				}

				output = append(output, msg)
			}
		}
	}

	return output, nil
}

func (t *procCapture) getStringMatch(match []string) string {
	if len(match) > 1 {
		return match[len(match)-1]
	}

	return ""
}
