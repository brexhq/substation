package transform

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

func newTimeFromUnix(_ context.Context, cfg config.Config) (*timeFromUnix, error) {
	conf := timeUnixConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform time_from_unix: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "time_from_unix"
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	tf := timeFromUnix{
		conf:     conf,
		isObject: conf.Object.SourceKey != "" && conf.Object.TargetKey != "",
	}

	return &tf, nil
}

type timeFromUnix struct {
	conf     timeUnixConfig
	isObject bool
}

func (tf *timeFromUnix) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
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

	// Convert Unix to UnixNano.
	date := time.Unix(value.Int(), 0)
	ns := date.UnixNano()

	if tf.isObject {
		if err := msg.SetValue(tf.conf.Object.TargetKey, ns); err != nil {
			return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
		}
	} else {
		value := []byte(fmt.Sprintf("%d", ns))
		msg.SetData(value)
	}

	return []*message.Message{msg}, nil
}

func (tf *timeFromUnix) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
