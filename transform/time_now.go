package transform

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/brexhq/substation/config"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/message"
)

type timeNowConfig struct {
	Object iconfig.Object `json:"object"`
}

func (c *timeNowConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *timeNowConfig) Validate() error {
	return nil
}

type timeNow struct {
	conf            timeNowConfig
	hasObjectSetKey bool
}

func newTimeNow(_ context.Context, cfg config.Config) (*timeNow, error) {
	conf := timeNowConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("time: new_now: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("time: new_now: %v", err)
	}

	tf := timeNow{
		conf:            conf,
		hasObjectSetKey: conf.Object.SetKey != "",
	}

	return &tf, nil
}

func (tf *timeNow) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	// Skip interrupt messages.
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	date := time.Now()

	if tf.hasObjectSetKey {
		if err := msg.SetValue(tf.conf.Object.SetKey, date.UnixMilli()); err != nil {
			return nil, fmt.Errorf("time: now: %v", err)
		}

		return []*message.Message{msg}, nil
	}

	value := timeUnixToBytes(date)
	msg.SetData(value)

	return []*message.Message{msg}, nil
}

func (tf *timeNow) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}

func (*timeNow) Close(context.Context) error {
	return nil
}
