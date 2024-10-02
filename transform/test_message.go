package transform

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"

	iconfig "github.com/brexhq/substation/v2/internal/config"
)

type testMessageConfig struct {
	Value interface{} `json:"value"`

	ID string `json:"id"`
}

func (c *testMessageConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func newTestMessage(_ context.Context, cfg config.Config) (*testMessage, error) {
	conf := testMessageConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform test_message: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "test_message"
	}

	tf := testMessage{
		conf: conf,
	}

	return &tf, nil
}

type testMessage struct {
	conf testMessageConfig
}

func (tf *testMessage) Transform(_ context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		m := message.New().SetData(anyToBytes(tf.conf.Value))
		return []*message.Message{m, msg}, nil
	}

	return []*message.Message{msg}, nil
}

func (tf *testMessage) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
