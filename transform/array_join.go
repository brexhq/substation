package transform

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/brexhq/substation/config"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/message"
)

type arrayJoinConfig struct {
	Object iconfig.Object `json:"object"`

	// Separator is the string that joins the array.
	Separator string `json:"separator"`
}

func (c *arrayJoinConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *arrayJoinConfig) Validate() error {
	if c.Object.Key == "" && c.Object.SetKey != "" {
		return fmt.Errorf("object_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.Object.Key != "" && c.Object.SetKey == "" {
		return fmt.Errorf("object_set_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.Separator == "" {
		return fmt.Errorf("separator: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

func newArrayJoin(_ context.Context, cfg config.Config) (*arrayJoin, error) {
	conf := arrayJoinConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: new_array_join: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: new_array_join: %v", err)
	}

	tf := arrayJoin{
		conf:            conf,
		hasObjectKey:    conf.Object.Key != "",
		hasObjectSetKey: conf.Object.SetKey != "",
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
		value = msg.GetValue(tf.conf.Object.Key)
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
		if err := msg.SetValue(tf.conf.Object.SetKey, str); err != nil {
			return nil, fmt.Errorf("transform: array_join: %v", err)
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
