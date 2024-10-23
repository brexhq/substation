package transform

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

func newTimeToUnix(_ context.Context, cfg config.Config) (*timeToUnix, error) {
	conf := timeUnixConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform time_to_unix: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "time_to_unix"
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	tf := timeToUnix{
		conf:     conf,
		isObject: conf.Object.SourceKey != "" && conf.Object.TargetKey != "",
	}

	return &tf, nil
}

type timeToUnix struct {
	conf     timeUnixConfig
	isObject bool
}

func (tf *timeToUnix) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
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

	// Convert UnixNano to Unix.
	date := time.Unix(0, value.Int())
	unix := date.Unix()

	if tf.isObject {
		if err := msg.SetValue(tf.conf.Object.TargetKey, unix); err != nil {
			return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
		}
	} else {
		value := []byte(fmt.Sprintf("%d", unix))
		msg.SetData(value)
	}

	return []*message.Message{msg}, nil
}

func (tf *timeToUnix) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
