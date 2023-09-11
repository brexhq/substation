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
	Object iconfig.Object `json:"object"`

	// Transform is the transform that is applied to each item in the array.
	Transform config.Config `json:"transform"`
}

func (c *metaForEachConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *metaForEachConfig) Validate() error {
	if c.Object.Key == "" {
		return fmt.Errorf("object_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.Object.SetKey == "" {
		return fmt.Errorf("object_set_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.Transform.Type == "" {
		return fmt.Errorf("type: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

type metaForEach struct {
	conf metaForEachConfig

	tf    Transformer
	tfCfg config.Config
}

func newMetaForEach(ctx context.Context, cfg config.Config) (*metaForEach, error) {
	conf := metaForEachConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: new_meta_for_each: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: new_meta_for_each: %v", err)
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
		return nil, fmt.Errorf("transform: new_meta_for_each: %v", err)
	}
	tf.tf = tfer

	return &tf, nil
}

func (tf *metaForEach) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	result := msg.GetValue(tf.conf.Object.Key)
	if !result.IsArray() {
		return []*message.Message{msg}, nil
	}

	var arr []interface{}
	for _, res := range result.Array() {
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

	if err := msg.SetValue(tf.conf.Object.SetKey, arr); err != nil {
		return nil, fmt.Errorf("transform: meta_for_each: %v", err)
	}

	return []*message.Message{msg}, nil
}

func (meta *metaForEach) String() string {
	b, _ := json.Marshal(meta.conf)
	return string(b)
}

func (*metaForEach) Close(context.Context) error {
	return nil
}
