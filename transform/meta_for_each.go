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

type metaForEachConfig struct {
	// Transform that is applied to each item in the array.
	Transform config.Config `json:"transform"`

	Object iconfig.Object `json:"object"`
}

func (c *metaForEachConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *metaForEachConfig) Validate() error {
	if c.Object.SourceKey == "" {
		return fmt.Errorf("object_source_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.Object.TargetKey == "" {
		return fmt.Errorf("object_target_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.Transform.Type == "" {
		return fmt.Errorf("type: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

func newMetaForEach(ctx context.Context, cfg config.Config) (*metaForEach, error) {
	conf := metaForEachConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: meta_for_each: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: meta_for_each: %v", err)
	}

	tf := metaForEach{
		conf: conf,
	}

	tfConf, err := json.Marshal(conf.Transform)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(tfConf, &tf.tfCfg); err != nil {
		return nil, err
	}

	tfer, err := New(ctx, tf.tfCfg)
	if err != nil {
		return nil, fmt.Errorf("transform: meta_for_each: %v", err)
	}
	tf.tf = tfer

	return &tf, nil
}

type metaForEach struct {
	conf metaForEachConfig

	tf    Transformer
	tfCfg config.Config
}

func (tf *metaForEach) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		msgs, err := tf.tf.Transform(ctx, msg)
		if err != nil {
			return nil, fmt.Errorf("transform: meta_for_each: %v", err)
		}

		msgs = append(msgs, msg)
		return msgs, nil
	}

	value := msg.GetValue(tf.conf.Object.SourceKey)
	if !value.Exists() {
		return []*message.Message{msg}, nil
	}

	if !value.IsArray() {
		return []*message.Message{msg}, nil
	}

	var arr []interface{}
	for _, res := range value.Array() {
		tmpMsg := message.New().SetData(res.Bytes())
		tfMsgs, err := tf.tf.Transform(ctx, tmpMsg)
		if err != nil {
			return nil, fmt.Errorf("transform: meta_for_each: %v", err)
		}

		for _, m := range tfMsgs {
			v := bytesToValue(m.Data())
			arr = append(arr, v.Value())
		}
	}

	if err := msg.SetValue(tf.conf.Object.TargetKey, arr); err != nil {
		return nil, fmt.Errorf("transform: meta_for_each: %v", err)
	}

	return []*message.Message{msg}, nil
}

func (tf *metaForEach) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
