package transform

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

func newTimeToString(_ context.Context, cfg config.Config) (*timeToString, error) {
	conf := timePatternConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform time_to_string: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "time_to_string"
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	tf := timeToString{
		conf:     conf,
		isObject: conf.Object.SourceKey != "" && conf.Object.TargetKey != "",
	}

	return &tf, nil
}

type timeToString struct {
	conf     timePatternConfig
	isObject bool
}

func (tf *timeToString) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.HasFlag(message.IsControl) {
		return []*message.Message{msg}, nil
	}

	var value message.Value
	if tf.isObject {
		value = msg.GetValue(tf.conf.Object.SourceKey)
	} else {
		value = bytesToValue(msg.Data())
	}

	if skipMessage(msg, value) {
		return []*message.Message{msg}, nil
	}

	pattern, err := timeUnixToStr(value.Int(), tf.conf.Format, tf.conf.Location)
	if err != nil {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
	}

	if tf.isObject {
		if err := msg.SetValue(tf.conf.Object.TargetKey, pattern); err != nil {
			return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
		}
	} else {
		msg.SetData([]byte(pattern))
	}

	return []*message.Message{msg}, nil
}

func (tf *timeToString) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
