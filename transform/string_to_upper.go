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

func newStringToUpper(_ context.Context, cfg config.Config) (*stringToUpper, error) {
	conf := strCaseConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: string_to_lower: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: string_to_lower: %v", err)
	}

	tf := stringToUpper{
		conf:     conf,
		isObject: conf.Object.SrcKey != "" && conf.Object.DstKey != "",
	}

	return &tf, nil
}

type stringToUpper struct {
	conf     strCaseConfig
	isObject bool
}

func (tf *stringToUpper) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	if !tf.isObject {
		b := bytes.ToUpper(msg.Data())
		msg.SetData(b)

		return []*message.Message{msg}, nil
	}

	value := msg.GetValue(tf.conf.Object.SrcKey)
	if !value.Exists() {
		return []*message.Message{msg}, nil
	}

	s := strings.ToUpper(value.String())
	if err := msg.SetValue(tf.conf.Object.DstKey, s); err != nil {
		return nil, fmt.Errorf("transform: string_to_lower: %v", err)
	}

	return []*message.Message{msg}, nil
}

func (tf *stringToUpper) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
