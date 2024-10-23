package transform

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/iancoleman/strcase"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

func newStringToSnake(_ context.Context, cfg config.Config) (*stringToSnake, error) {
	conf := strCaseConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform string_to_snake: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "string_to_snake"
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	tf := stringToSnake{
		conf:     conf,
		isObject: conf.Object.SourceKey != "" && conf.Object.TargetKey != "",
	}

	return &tf, nil
}

type stringToSnake struct {
	conf     strCaseConfig
	isObject bool
}

func (tf *stringToSnake) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.HasFlag(message.IsControl) {
		return []*message.Message{msg}, nil
	}

	if !tf.isObject {
		b := []byte(strcase.ToSnake(string(msg.Data())))
		msg.SetData(b)

		return []*message.Message{msg}, nil
	}

	value := msg.GetValue(tf.conf.Object.SourceKey)
	if skipMessage(msg, value) {
		return []*message.Message{msg}, nil
	}

	s := strcase.ToSnake(value.String())
	if err := msg.SetValue(tf.conf.Object.TargetKey, s); err != nil {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
	}

	return []*message.Message{msg}, nil
}

func (tf *stringToSnake) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
