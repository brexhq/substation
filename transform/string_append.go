package transform

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"

	iconfig "github.com/brexhq/substation/v2/internal/config"
)

type stringAppendConfig struct {
	// Suffix is the string appended to the end of the string.
	Suffix string `json:"suffix"`

	ID     string         `json:"id"`
	Object iconfig.Object `json:"object"`
}

func (c *stringAppendConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *stringAppendConfig) Validate() error {
	if c.Object.SourceKey == "" && c.Object.TargetKey != "" {
		return fmt.Errorf("object_source_key: %v", iconfig.ErrMissingRequiredOption)
	}

	if c.Object.SourceKey != "" && c.Object.TargetKey == "" {
		return fmt.Errorf("object_target_key: %v", iconfig.ErrMissingRequiredOption)
	}

	if c.Suffix == "" {
		return fmt.Errorf("suffix: %v", iconfig.ErrMissingRequiredOption)
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
		return nil, fmt.Errorf("transform string_append: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "string_append"
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	tf := stringAppend{
		conf:     conf,
		isObject: conf.Object.SourceKey != "" && conf.Object.TargetKey != "",
		s:        []byte(conf.Suffix),
	}

	return &tf, nil
}

func (tf *stringAppend) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.HasFlag(message.IsControl) {
		return []*message.Message{msg}, nil
	}

	if !tf.isObject {
		b := msg.Data()
		b = append(b, tf.s...)

		msg.SetData(b)
		return []*message.Message{msg}, nil
	}

	value := msg.GetValue(tf.conf.Object.SourceKey)
	if skipMessage(msg, value) {
		return []*message.Message{msg}, nil
	}

	str := value.String() + tf.conf.Suffix

	if err := msg.SetValue(tf.conf.Object.TargetKey, str); err != nil {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
	}

	return []*message.Message{msg}, nil
}

func (tf *stringAppend) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
