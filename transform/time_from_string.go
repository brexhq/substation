package transform

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

func newTimeFromString(_ context.Context, cfg config.Config) (*timeFromString, error) {
	conf := timePatternConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform time_from_string: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "time_from_string"
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	tf := timeFromString{
		conf:     conf,
		isObject: conf.Object.SourceKey != "" && conf.Object.TargetKey != "",
	}

	return &tf, nil
}

type timeFromString struct {
	conf     timePatternConfig
	isObject bool
}

func (tf *timeFromString) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	var value message.Value
	if tf.isObject {
		value = msg.GetValue(tf.conf.Object.SourceKey)
	} else {
		value = bytesToValue(msg.Data())
	}

	if !value.Exists() {
		return []*message.Message{msg}, nil
	}

	date, err := timeStrToUnix(value.String(), tf.conf.Format, tf.conf.Location)
	if err != nil {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
	}

	if tf.isObject {
		if err := msg.SetValue(tf.conf.Object.TargetKey, date.UnixNano()); err != nil {
			return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
		}
	} else {
		value := timeUnixToBytes(date)
		msg.SetData(value)
	}

	return []*message.Message{msg}, nil
}

func (tf *timeFromString) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
