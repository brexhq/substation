package transform

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

func newStringToLower(_ context.Context, cfg config.Config) (*stringToLower, error) {
	conf := strCaseConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform string_to_lower: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "string_to_lower"
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	tf := stringToLower{
		conf:     conf,
		isObject: conf.Object.SourceKey != "" && conf.Object.TargetKey != "",
	}

	return &tf, nil
}

type stringToLower struct {
	conf     strCaseConfig
	isObject bool
}

func (tf *stringToLower) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.HasFlag(message.IsControl) {
		return []*message.Message{msg}, nil
	}

	if !tf.isObject {
		b := bytes.ToLower(msg.Data())
		msg.SetData(b)

		return []*message.Message{msg}, nil
	}

	value := msg.GetValue(tf.conf.Object.SourceKey)
	if skipMessage(msg, value) {
		return []*message.Message{msg}, nil
	}

	s := strings.ToLower(value.String())
	if err := msg.SetValue(tf.conf.Object.TargetKey, s); err != nil {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
	}

	return []*message.Message{msg}, nil
}

func (tf *stringToLower) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
