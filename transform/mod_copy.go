package transform

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/message"
)

type modCopyConfig struct {
	Object configObject `json:"object"`
}

type modCopy struct {
	conf      modCopyConfig
	isFromObj bool
	isToObj   bool
}

func newModCopy(_ context.Context, cfg config.Config) (*modCopy, error) {
	conf := modCopyConfig{}
	if err := iconfig.Decode(cfg.Settings, &conf); err != nil {
		return nil, fmt.Errorf("transform: new_mod_copy: %v", err)
	}

	tf := modCopy{
		conf:      conf,
		isFromObj: conf.Object.Key != "" && conf.Object.SetKey == "",
		isToObj:   conf.Object.Key == "" && conf.Object.SetKey != "",
	}

	return &tf, nil
}

func (tf *modCopy) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}

func (*modCopy) Close(context.Context) error {
	return nil
}

func (tf *modCopy) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	// Skip interrupt messages.
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	if tf.isFromObj {
		res := msg.GetObject(tf.conf.Object.Key)
		outMsg := message.New().SetData(res.Bytes()).SetMetadata(msg.Metadata())

		return []*message.Message{outMsg}, nil
	}

	if tf.isToObj {
		outMsg := message.New().SetMetadata(msg.Metadata())
		if err := outMsg.SetObject(tf.conf.Object.SetKey, msg.Data()); err != nil {
			return nil, fmt.Errorf("transform: mod_copy: %v", err)
		}

		return []*message.Message{outMsg}, nil
	}

	if err := msg.SetObject(tf.conf.Object.SetKey, msg.GetObject(tf.conf.Object.Key)); err != nil {
		return nil, fmt.Errorf("transform: mod_copy: %v", err)
	}

	return []*message.Message{msg}, nil
}
