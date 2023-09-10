package transform

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
	"github.com/iancoleman/strcase"
)

func newStringToSnake(_ context.Context, cfg config.Config) (*stringToSnake, error) {
	conf := strCaseConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: new_str_case_snake: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: new_str_case_snake: %v", err)
	}

	tf := stringToSnake{
		conf:     conf,
		isObject: conf.Object.Key != "" && conf.Object.SetKey != "",
	}

	return &tf, nil
}

type stringToSnake struct {
	conf     strCaseConfig
	isObject bool
}

func (tf *stringToSnake) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	if !tf.isObject {
		b := []byte(strcase.ToSnake(string(msg.Data())))
		msg.SetData(b)

		return []*message.Message{msg}, nil
	}

	v := msg.GetValue(tf.conf.Object.Key)
	s := strcase.ToSnake(v.String())

	if err := msg.SetValue(tf.conf.Object.SetKey, s); err != nil {
		return nil, fmt.Errorf("transform: str_case_snake: %v", err)
	}

	return []*message.Message{msg}, nil
}

func (tf *stringToSnake) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}

func (*stringToSnake) Close(context.Context) error {
	return nil
}