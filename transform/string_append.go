package transform

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/message"
)

type stringAppendConfig struct {
	Object iconfig.Object `json:"object"`

	// Suffix is the string to append.
	Suffix string `json:"suffix"`
}

func (c *stringAppendConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *stringAppendConfig) Validate() error {
	if c.Object.SrcKey == "" && c.Object.DstKey != "" {
		return fmt.Errorf("object_src_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.Object.SrcKey != "" && c.Object.DstKey == "" {
		return fmt.Errorf("object_dst_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.Suffix == "" {
		return fmt.Errorf("suffix: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

type stringAppend struct {
	conf     stringAppendConfig
	isObject bool

	s []byte
}

func newStringAppend(_ context.Context, cfg config.Config) (*stringAppend, error) {
	conf := stringAppendConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: string_append: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: string_append: %v", err)
	}

	tf := stringAppend{
		conf:     conf,
		isObject: conf.Object.SrcKey != "" && conf.Object.DstKey != "",
		s:        []byte(conf.Suffix),
	}

	return &tf, nil
}

func (tf *stringAppend) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	if !tf.isObject {
		b := msg.Data()
		b = append(b, tf.s...)

		msg.SetData(b)
		return []*message.Message{msg}, nil
	}

	value := msg.GetValue(tf.conf.Object.SrcKey)
	if !value.Exists() {
		return []*message.Message{msg}, nil
	}

	str := value.String() + tf.conf.Suffix

	if err := msg.SetValue(tf.conf.Object.DstKey, str); err != nil {
		return nil, fmt.Errorf("transform: string_append: %v", err)
	}

	return []*message.Message{msg}, nil
}

func (tf *stringAppend) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
