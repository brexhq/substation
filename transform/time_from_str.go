package transform

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

// timeFromStr is a transform that converts a pattern-based string
// format to a UnixMilli timestamp.
type timeFromStr struct {
	conf     timePatternConfig
	isObject bool
}

func newTimeFromStr(_ context.Context, cfg config.Config) (*timeFromStr, error) {
	conf := timePatternConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("time: new_from_string: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("time: new_from_string: %v", err)
	}

	tf := timeFromStr{
		conf:     conf,
		isObject: conf.Object.Key != "" && conf.Object.SetKey != "",
	}

	return &tf, nil
}

func (tf *timeFromStr) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	var value message.Value
	if tf.isObject {
		value = msg.GetValue(tf.conf.Object.Key)
	} else {
		value = bytesToValue(msg.Data())
	}

	if !value.Exists() {
		return []*message.Message{msg}, nil
	}

	date, err := timeStrToUnix(value.String(), tf.conf.Format, tf.conf.Location)
	if err != nil {
		return nil, fmt.Errorf("transform: time_from_str: %v", err)
	}

	if tf.isObject {
		if err := msg.SetValue(tf.conf.Object.SetKey, date.UnixMilli()); err != nil {
			return nil, fmt.Errorf("transform: time_from_str: %v", err)
		}
	} else {
		value := timeUnixToBytes(date)
		msg.SetData(value)
	}

	return []*message.Message{msg}, nil
}

func (tf *timeFromStr) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}

func (*timeFromStr) Close(context.Context) error {
	return nil
}
