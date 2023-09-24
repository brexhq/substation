package transform

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/message"
	"github.com/google/uuid"
)

type stringUUIDConfig struct {
	Object iconfig.Object `json:"object"`
}

func (c *stringUUIDConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func newStringUUID(_ context.Context, cfg config.Config) (*stringUUID, error) {
	conf := stringUUIDConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: string_uuid: %v", err)
	}

	tf := stringUUID{
		conf:            conf,
		hasObjectSetKey: conf.Object.SetKey != "",
	}

	return &tf, nil
}

type stringUUID struct {
	conf            stringUUIDConfig
	hasObjectSetKey bool
}

func (tf *stringUUID) Transform(_ context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	uid := uuid.NewString()
	if tf.hasObjectSetKey {
		if err := msg.SetValue(tf.conf.Object.SetKey, uid); err != nil {
			return nil, fmt.Errorf("transform: string_uuid: %v", err)
		}

		return []*message.Message{msg}, nil
	}

	msg.SetData([]byte(uid))
	return []*message.Message{msg}, nil
}

func (tf *stringUUID) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
