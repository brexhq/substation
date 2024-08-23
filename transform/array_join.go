package transform

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/brexhq/substation/v2/config"
	iconfig "github.com/brexhq/substation/v2/internal/config"
	"github.com/brexhq/substation/v2/internal/errors"
	"github.com/brexhq/substation/v2/message"
)

type arrayJoinConfig struct {
	// Separator is the string that is used to join data.
	Separator string `json:"separator"`

	ID     string         `json:"id"`
	Object iconfig.Object `json:"object"`
}

func (c *arrayJoinConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *arrayJoinConfig) Validate() error {
	if c.Object.SourceKey == "" && c.Object.TargetKey != "" {
		return fmt.Errorf("object_source_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.Object.SourceKey != "" && c.Object.TargetKey == "" {
		return fmt.Errorf("object_target_key: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

func newArrayJoin(_ context.Context, cfg config.Config) (*arrayJoin, error) {
	conf := arrayJoinConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform array_join: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "array_join"
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	tf := arrayJoin{
		conf:            conf,
		hasObjectKey:    conf.Object.SourceKey != "",
		hasObjectSetKey: conf.Object.TargetKey != "",
		separator:       []byte(conf.Separator),
	}

	return &tf, nil
}

type arrayJoin struct {
	conf            arrayJoinConfig
	hasObjectKey    bool
	hasObjectSetKey bool

	separator []byte
}

func (tf *arrayJoin) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	var value message.Value
	if tf.hasObjectKey {
		value = msg.GetValue(tf.conf.Object.SourceKey)
	} else {
		value = bytesToValue(msg.Data())
	}

	if !value.Exists() {
		return []*message.Message{msg}, nil
	}

	if !value.IsArray() {
		return []*message.Message{msg}, nil
	}

	var arr []string
	for _, val := range value.Array() {
		arr = append(arr, val.String())
	}

	str := strings.Join(arr, tf.conf.Separator)

	if tf.hasObjectSetKey {
		if err := msg.SetValue(tf.conf.Object.TargetKey, str); err != nil {
			return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
		}

		return []*message.Message{msg}, nil
	}

	msg.SetData([]byte(str))
	return []*message.Message{msg}, nil
}

func (tf *arrayJoin) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
