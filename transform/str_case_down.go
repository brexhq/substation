package transform

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/brexhq/substation/config"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/message"
)

type strCaseDownConfig struct {
	Object iconfig.Object `json:"object"`
}

func (c *strCaseDownConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *strCaseDownConfig) Validate() error {
	if c.Object.Key == "" && c.Object.SetKey != "" {
		return fmt.Errorf("object_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.Object.Key != "" && c.Object.SetKey == "" {
		return fmt.Errorf("object_set_key: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

type strCaseDown struct {
	conf     strCaseDownConfig
	isObject bool
}

func newStrCaseDown(_ context.Context, cfg config.Config) (*strCaseDown, error) {
	conf := strCaseDownConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: new_str_case_down: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: new_str_case_down: %v", err)
	}

	tf := strCaseDown{
		conf:     conf,
		isObject: conf.Object.Key != "" && conf.Object.SetKey != "",
	}

	return &tf, nil
}

func (tf *strCaseDown) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	// Skip interrupt messages.
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	if !tf.isObject {
		b := bytes.ToLower(msg.Data())
		msg.SetData(b)

		return []*message.Message{msg}, nil
	}

	v := msg.GetValue(tf.conf.Object.Key).String()
	s := strings.ToLower(v)

	if err := msg.SetValue(tf.conf.Object.SetKey, s); err != nil {
		return nil, fmt.Errorf("transform: str_case_down: %v", err)
	}

	return []*message.Message{msg}, nil
}

func (tf *strCaseDown) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}

func (*strCaseDown) Close(context.Context) error {
	return nil
}
