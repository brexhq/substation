package transform

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/message"
)

type utilityDropConfig struct{}

func (c *utilityDropConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func newUtilityDrop(_ context.Context, cfg config.Config) (*utilityDrop, error) {
	conf := utilityDropConfig{}
	if err := iconfig.Decode(cfg.Settings, &conf); err != nil {
		return nil, fmt.Errorf("transform: new_util_drop: %v", err)
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
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	return []*message.Message{}, nil
}

func (tf *utilityDrop) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}

func (*utilityDrop) Close(context.Context) error {
	return nil
}
