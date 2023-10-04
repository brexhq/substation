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
		return nil, fmt.Errorf("transform: time_from_unix: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: time_from_unix: %v", err)
	}

	tf := timeFromUnix{
		conf:     conf,
		isObject: conf.Object.Key != "" && conf.Object.SetKey != "",
	}

	return &tf, nil
}

// TimeFromUnix is a transform that converts a UnixMilli timestamp to a
// Unix timestamp.
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
		value = msg.GetValue(tf.conf.Object.Key)
	} else {
		value = bytesToValue(msg.Data())
	}

	if !value.Exists() {
		return []*message.Message{msg}, nil
	}

	// Convert Unix to UnixMilli.
	date := time.Unix(value.Int(), 0)
	milli := date.UnixMilli()

	if tf.isObject {
		if err := msg.SetValue(tf.conf.Object.SetKey, milli); err != nil {
			return nil, fmt.Errorf("transform: time_from_unix: %v", err)
		}
	} else {
		value := []byte(fmt.Sprintf("%d", milli))
		msg.SetData(value)
	}

	return []*message.Message{msg}, nil
}

func (tf *timeFromUnix) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
