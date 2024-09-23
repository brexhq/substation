package transform

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

func newHashSHA256(_ context.Context, cfg config.Config) (*hashSHA256, error) {
	conf := hashConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform hash_sha256: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "hash_sha256"
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	tf := hashSHA256{
		conf:     conf,
		isObject: conf.Object.SourceKey != "" && conf.Object.TargetKey != "",
	}

	return &tf, nil
}

type hashSHA256 struct {
	conf     hashConfig
	isObject bool
}

func (tf *hashSHA256) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	if !tf.isObject {
		sum := sha256.Sum256(msg.Data())
		str := fmt.Sprintf("%x", sum)

		msg.SetData([]byte(str))
		return []*message.Message{msg}, nil
	}

	value := msg.GetValue(tf.conf.Object.SourceKey)
	if !value.Exists() {
		return []*message.Message{msg}, nil
	}

	sum := sha256.Sum256(value.Bytes())
	str := fmt.Sprintf("%x", sum)

	if err := msg.SetValue(tf.conf.Object.TargetKey, str); err != nil {
		return nil, err
	}

	return []*message.Message{msg}, nil
}

func (tf *hashSHA256) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
