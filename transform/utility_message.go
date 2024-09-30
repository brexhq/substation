package transform

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"

	iconfig "github.com/brexhq/substation/v2/internal/config"
)

type utilityMessageConfig struct {
	Value interface{} `json:"value"`

	ID string `json:"id"`
}

func (c *utilityMessageConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func newUtilityMessage(_ context.Context, cfg config.Config) (*utilityMessage, error) {
	conf := utilityMessageConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform utility_message: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "utility_message"
	}

	tf := utilityMessage{
		conf: conf,
	}

	return &tf, nil
}

type utilityMessage struct {
	conf utilityMessageConfig
}

func (tf *utilityMessage) Transform(_ context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		m := message.New().SetData(anyToBytes(tf.conf.Value))
		return []*message.Message{m, msg}, nil
	}

	return []*message.Message{msg}, nil
}

func (tf *utilityMessage) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
