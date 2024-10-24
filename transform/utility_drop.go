package transform

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"

	iconfig "github.com/brexhq/substation/v2/internal/config"
)

type utilityDropConfig struct {
	ID string `json:"id"`
}

func (c *utilityDropConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func newUtilityDrop(_ context.Context, cfg config.Config) (*utilityDrop, error) {
	conf := utilityDropConfig{}
	if err := iconfig.Decode(cfg.Settings, &conf); err != nil {
		return nil, fmt.Errorf("transform utility_drop: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "utility_drop"
	}

	tf := utilityDrop{
		conf: conf,
	}

	return &tf, nil
}

type utilityDrop struct {
	conf utilityDropConfig
}

func (tf *utilityDrop) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.HasFlag(message.IsControl) {
		return []*message.Message{msg}, nil
	}

	return []*message.Message{}, nil
}

func (tf *utilityDrop) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
