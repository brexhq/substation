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

type strCaptureFindConfig struct {
	Object iconfig.Object `json:"object"`

	// Expression is the regular expression used to capture values.
	Expression string `json:"expression"`

	re *regexp.Regexp
}

func (c *strCaptureFindConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *strCaptureFindConfig) Validate() error {
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

	return nil
}

type strCaptureFind struct {
	conf     strCaptureFindConfig
	isObject bool
}

func newStrCaptureFind(_ context.Context, cfg config.Config) (*strCaptureFind, error) {
	conf := strCaptureFindConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: new_str_capture_find: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: new_str_capture_find: %v", err)
	}

	tf := strCaptureFind{
		conf:     conf,
		isObject: conf.Object.Key != "" && conf.Object.SetKey != "",
	}

	return &tf, nil
}

func (tf *strCaptureFind) Transform(_ context.Context, msg *message.Message) ([]*message.Message, error) {
	// Skip Capture messages.
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	if !tf.isObject {
		matches := tf.conf.re.FindSubmatch(msg.Data())
		msg.SetData(captureGetBytesMatch(matches))

		return []*message.Message{msg}, nil
	}

	v := msg.GetValue(tf.conf.Object.Key)
	matches := tf.conf.re.FindStringSubmatch(v.String())

	if err := msg.SetValue(tf.conf.Object.SetKey, captureGetStringMatch(matches)); err != nil {
		return nil, fmt.Errorf("transform: str_capture_find: %v", err)
	}

	return []*message.Message{msg}, nil
}

func (tf *strCaptureFind) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}

func (*strCaptureFind) Close(context.Context) error {
	return nil
}
