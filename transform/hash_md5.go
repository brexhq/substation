package transform

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

func newHashMD5(_ context.Context, cfg config.Config) (*hashMD5, error) {
	conf := hashConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: new_mod_hash: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: new_mod_hash: %v", err)
	}

	tf := hashMD5{
		conf:     conf,
		isObject: conf.Object.Key != "" && conf.Object.SetKey != "",
	}

	return &tf, nil
}

type hashMD5 struct {
	conf     hashConfig
	isObject bool
}

func (tf *hashMD5) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	if !tf.isObject {
		sum := md5.Sum(msg.Data())
		str := fmt.Sprintf("%x", sum)

		msg.SetData([]byte(str))
		return []*message.Message{msg}, nil
	}

	value := msg.GetValue(tf.conf.Object.Key)
	sum := md5.Sum(value.Bytes())
	str := fmt.Sprintf("%x", sum)

	if err := msg.SetValue(tf.conf.Object.SetKey, str); err != nil {
		return nil, err
	}

	return []*message.Message{msg}, nil
}

func (tf *hashMD5) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
