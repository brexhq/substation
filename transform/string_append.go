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
	// Suffix is the static string to append. This is used when SuffixKey is empty.
	Suffix string `json:"suffix"`
	// SuffixKey is the object key to get the suffix value from. When set, this takes precedence over Suffix.
	SuffixKey string `json:"suffix_key"`

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

	if c.Suffix == "" && c.SuffixKey == "" {
		return fmt.Errorf("either suffix or suffix_key must be set")
	}

	// If SuffixKey is set, we need Object.SourceKey to be set as well
	if c.SuffixKey != "" && c.Object.SourceKey == "" {
		return fmt.Errorf("object_source_key is required when using suffix_key")
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
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	// Get the suffix value - either from SuffixKey or static Suffix
	var suffixStr string
	if tf.conf.SuffixKey != "" {
		suffixValue := msg.GetValue(tf.conf.SuffixKey)
		if !suffixValue.Exists() {
			return []*message.Message{msg}, nil
		}
		suffixStr = suffixValue.String()
	} else {
		suffixStr = string(tf.s)
	}

	if !tf.isObject {
		b := msg.Data()
		b = append(b, []byte(suffixStr)...)

		msg.SetData(b)
		return []*message.Message{msg}, nil
	}

	value := msg.GetValue(tf.conf.Object.SourceKey)
	if !value.Exists() {
		return []*message.Message{msg}, nil
	}

	str := value.String() + suffixStr

	if err := msg.SetValue(tf.conf.Object.TargetKey, str); err != nil {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
	}

	return []*message.Message{msg}, nil
}

func (tf *stringAppend) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
