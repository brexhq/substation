package transform

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

func newTimeToStr(_ context.Context, cfg config.Config) (*timeToStr, error) {
	conf := timePatternConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: new_time_to_str: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: new_time_to_str: %v", err)
	}

	tf := timeToStr{
		conf:     conf,
		isObject: conf.Object.Key != "" && conf.Object.SetKey != "",
	}

	return &tf, nil
}

// timeToStr is a transform that converts a Unix timestamp to a
// pattern-based string format.
type timeToStr struct {
	conf     timePatternConfig
	isObject bool
}

func (tf *timeToStr) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
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

	pattern, err := timeUnixToStr(value.Int(), tf.conf.Format, tf.conf.Location)
	if err != nil {
		return nil, fmt.Errorf("transform: time_to_str: %v", err)
	}

	if tf.isObject {
		if err := msg.SetValue(tf.conf.Object.SetKey, pattern); err != nil {
			return nil, fmt.Errorf("transform: time_to_str: %v", err)
		}
	} else {
		msg.SetData([]byte(pattern))
	}

	return []*message.Message{msg}, nil
}

func (tf *timeToStr) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}

func (*timeToStr) Close(context.Context) error {
	return nil
}
