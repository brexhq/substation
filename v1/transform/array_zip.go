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

type arrayZipConfig struct {
	ID     string         `json:"id"`
	Object iconfig.Object `json:"object"`
}

func (c *arrayZipConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *arrayZipConfig) Validate() error {
	if c.Object.SourceKey == "" && c.Object.TargetKey != "" {
		return fmt.Errorf("object_source_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.Object.SourceKey != "" && c.Object.TargetKey == "" {
		return fmt.Errorf("object_target_key: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

func newArrayZip(_ context.Context, cfg config.Config) (*arrayZip, error) {
	conf := arrayZipConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform array_zip: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "array_zip"
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	tf := arrayZip{
		conf:      conf,
		hasObjSrc: conf.Object.SourceKey != "",
		hasObjDst: conf.Object.TargetKey != "",
	}

	return &tf, nil
}

type arrayZip struct {
	conf      arrayZipConfig
	hasObjSrc bool
	hasObjDst bool
}

func (tf *arrayZip) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	var value message.Value
	if tf.hasObjSrc {
		value = msg.GetValue(tf.conf.Object.SourceKey)
	} else {
		value = bytesToValue(msg.Data())
	}

	if !value.Exists() || !value.IsArray() {
		return []*message.Message{msg}, nil
	}

	cache := make(map[int][]interface{})
	for _, val := range value.Array() {
		for i, v := range val.Array() {
			cache[i] = append(cache[i], v.Value())
		}
	}

	var b []interface{}
	for i := 0; i < len(cache); i++ {
		b = append(b, cache[i])
	}

	if tf.hasObjDst {
		if err := msg.SetValue(tf.conf.Object.TargetKey, b); err != nil {
			return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
		}

		return []*message.Message{msg}, nil
	}

	msg.SetData(anyToBytes(b))
	return []*message.Message{msg}, nil
}

func (tf *arrayZip) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
