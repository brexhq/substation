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
		isObject: conf.Object.SrcKey != "" && conf.Object.DstKey != "",
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
		value = msg.GetValue(tf.conf.Object.SrcKey)
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
		if err := msg.SetValue(tf.conf.Object.DstKey, ns); err != nil {
			return nil, fmt.Errorf("transform: time_from_unix: %v", err)
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
