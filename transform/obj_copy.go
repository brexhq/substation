package transform

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/message"
)

type objCopyConfig struct {
	Object iconfig.Object `json:"object"`
}

type objCopy struct {
	conf            objCopyConfig
	hasObjectKey    bool
	hasObjectSetKey bool
}

func newObjCopy(_ context.Context, cfg config.Config) (*objCopy, error) {
	conf := objCopyConfig{}
	if err := iconfig.Decode(cfg.Settings, &conf); err != nil {
		return nil, fmt.Errorf("transform: new_object_copy: %v", err)
	}

	tf := objCopy{
		conf:            conf,
		hasObjectKey:    conf.Object.Key != "" && conf.Object.SetKey == "",
		hasObjectSetKey: conf.Object.Key == "" && conf.Object.SetKey != "",
	}

	return &tf, nil
}

func (tf *objCopy) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	// Skip interrupt messages.
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	if tf.hasObjectKey {
		value := msg.GetValue(tf.conf.Object.Key)
		if !value.Exists() {
			return []*message.Message{msg}, nil
		}

		msg.SetData(value.Bytes())
		return []*message.Message{msg}, nil
	}

	if tf.hasObjectSetKey {
		if len(msg.Data()) == 0 {
			return []*message.Message{msg}, nil
		}

		outMsg := message.New().SetMetadata(msg.Metadata())
		if err := outMsg.SetValue(tf.conf.Object.SetKey, msg.Data()); err != nil {
			return nil, fmt.Errorf("transform: object_copy: %v", err)
		}

		return []*message.Message{outMsg}, nil
	}

	value := msg.GetValue(tf.conf.Object.Key)
	if !value.Exists() {
		return []*message.Message{msg}, nil
	}

	if err := msg.SetValue(tf.conf.Object.SetKey, value); err != nil {
		return nil, fmt.Errorf("transform: object_copy: %v", err)
	}

	return []*message.Message{msg}, nil
}

func (tf *objCopy) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}

func (*objCopy) Close(context.Context) error {
	return nil
}
