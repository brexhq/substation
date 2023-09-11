package transform

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

func newStringToLower(_ context.Context, cfg config.Config) (*stringToLower, error) {
	conf := strCaseConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: new_str_case_down: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: new_str_case_down: %v", err)
	}

	tf := stringToLower{
		conf:     conf,
		isObject: conf.Object.Key != "" && conf.Object.SetKey != "",
	}

	return &tf, nil
}

type stringToLower struct {
	conf     strCaseConfig
	isObject bool
}

func (tf *stringToLower) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
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

func (tf *stringToLower) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
