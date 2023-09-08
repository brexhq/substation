package transform

import (
	"context"
	gojson "encoding/json"
	"fmt"
	"strings"

	"github.com/brexhq/substation/config"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/message"
)

type arrJoinConfig struct {
	Object iconfig.Object `json:"object"`

	// Separator is the string that joins the array.
	Separator string `json:"separator"`
}

func (c *arrJoinConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *arrJoinConfig) Validate() error {
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

type arrJoin struct {
	conf            arrJoinConfig
	hasObjectKey    bool
	hasObjectSetKey bool

	separator []byte
}

func newArrJoin(_ context.Context, cfg config.Config) (*arrJoin, error) {
	conf := arrJoinConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: new_array_join: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: new_array_join: %v", err)
	}

	tf := arrJoin{
		conf:            conf,
		hasObjectKey:    conf.Object.Key != "",
		hasObjectSetKey: conf.Object.SetKey != "",
		separator:       []byte(conf.Separator),
	}

	return &tf, nil
}

func (tf *arrJoin) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	// Skip interrupt messages.
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

func (tf *arrJoin) String() string {
	b, _ := gojson.Marshal(tf.conf)
	return string(b)
}

func (*arrJoin) Close(context.Context) error {
	return nil
}
