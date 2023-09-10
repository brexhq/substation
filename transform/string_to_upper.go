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
		return nil, fmt.Errorf("strings: new_case_up: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("strings: new_case_up: %v", err)
	}

	tf := stringToUpper{
		conf:     conf,
		isObject: conf.Object.Key != "" && conf.Object.SetKey != "",
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

	v := msg.GetValue(tf.conf.Object.Key)
	s := strings.ToUpper(v.String())

	if err := msg.SetValue(tf.conf.Object.SetKey, s); err != nil {
		return nil, fmt.Errorf("transform: strings_case_up: %v", err)
	}

	return []*message.Message{msg}, nil
}

func (tf *stringToUpper) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}

func (*stringToUpper) Close(context.Context) error {
	return nil
}
