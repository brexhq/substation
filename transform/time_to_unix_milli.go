package transform

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

func newTimeToUnixMilli(_ context.Context, cfg config.Config) (*timeToUnixMilli, error) {
	conf := timeUnixConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform time_to_unix_milli: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "time_to_unix_milli"
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	tf := timeToUnixMilli{
		conf:     conf,
		isObject: conf.Object.SourceKey != "" && conf.Object.TargetKey != "",
	}

	return &tf, nil
}

type timeToUnixMilli struct {
	conf     timeUnixConfig
	isObject bool
}

func (tf *timeToUnixMilli) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
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

	// Convert UnixNano to UnixMilli.
	date := time.Unix(0, value.Int())
	ms := date.UnixMilli()

	if tf.isObject {
		if err := msg.SetValue(tf.conf.Object.TargetKey, ms); err != nil {
			return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
		}
	} else {
		value := []byte(fmt.Sprintf("%d", ms))
		msg.SetData(value)
	}

	return []*message.Message{msg}, nil
}

func (tf *timeToUnixMilli) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
