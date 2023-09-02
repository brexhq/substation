package transform

import (
	"context"
	gojson "encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/message"
)

type metaForEachConfig struct {
	Object configObject `json:"object"`

	// Transform is the transform that is applied to each item in the array.
	Transform config.Config `json:"transform"`
}

type metaForEach struct {
	conf metaForEachConfig

	tf    Transformer
	tfCfg config.Config
}

func newMetaForEach(ctx context.Context, cfg config.Config) (*metaForEach, error) {
	conf := metaForEachConfig{}
	if err := iconfig.Decode(cfg.Settings, &conf); err != nil {
		return nil, fmt.Errorf("transform: new_meta_for_each: %v", err)
	}

	// Validate required options.
	if conf.Object.Key == "" {
		return nil, fmt.Errorf("transform: new_meta_for_each: object_key: %v", errors.ErrMissingRequiredOption)
	}

	if conf.Object.SetKey == "" {
		return nil, fmt.Errorf("transform: new_meta_for_each: object_set_key: %v", errors.ErrMissingRequiredOption)
	}

	if conf.Transform.Type == "" {
		return nil, fmt.Errorf("transform: new_meta_for_each: type: %v", errors.ErrMissingRequiredOption)
	}

	tf := metaForEach{
		conf: conf,
	}

	tfConf, err := gojson.Marshal(conf.Transform)
	if err != nil {
		return nil, err
	}

	if err := gojson.Unmarshal(tfConf, &tf.tfCfg); err != nil {
		return nil, err
	}

	tfer, err := New(ctx, tf.tfCfg)
	if err != nil {
		return nil, fmt.Errorf("transform: new_meta_for_each: %v", err)
	}
	tf.tf = tfer

	return &tf, nil
}

func (meta *metaForEach) String() string {
	b, _ := gojson.Marshal(meta.conf)
	return string(b)
}

func (*metaForEach) Close(context.Context) error {
	return nil
}

func (tf *metaForEach) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	result := msg.GetObject(tf.conf.Object.Key)
	if !result.IsArray() {
		return []*message.Message{msg}, nil
	}

	var arr []interface{}
	for _, res := range result.Array() {
		tempMsg := message.New().SetData(res.Bytes())
		tfMsgs, err := tf.tf.Transform(ctx, tempMsg)
		if err != nil {
			return nil, fmt.Errorf("transform: meta_for_each: %v", err)
		}

		for _, m := range tfMsgs {
			tmp := message.New()
			if err := tmp.SetObject(tf.tfCfg.Type, m.Data()); err != nil {
				return nil, fmt.Errorf("transform: meta_for_each: %v", err)
			}

			arr = append(arr, tmp.GetObject(tf.tfCfg.Type).Value())
		}
	}

	if err := msg.SetObject(tf.conf.Object.SetKey, arr); err != nil {
		return nil, fmt.Errorf("transform: meta_for_each: %v", err)
	}

	return []*message.Message{msg}, nil
}
