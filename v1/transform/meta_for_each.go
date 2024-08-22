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
	//
	// Deprecated: Transform exists for backwards compatibility and will be
	// removed in a future release. Use Transforms instead.
	Transform config.Config `json:"transform"`
	// Transforms that are applied in series to the data in the array.
	Transforms []config.Config

	ID     string         `json:"id"`
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

	if c.Transform.Type == "" && len(c.Transforms) == 0 {
		return fmt.Errorf("type: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

func newMetaForEach(ctx context.Context, cfg config.Config) (*metaForEach, error) {
	conf := metaForEachConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform meta_for_each: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "meta_for_each"
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	tf := metaForEach{
		conf: conf,
	}

	if conf.Transform.Type != "" {
		tfer, err := New(ctx, conf.Transform)
		if err != nil {
			return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
		}
		tf.tf = tfer
	}

	tf.tfs = make([]Transformer, len(conf.Transforms))
	for i, t := range conf.Transforms {
		tfer, err := New(ctx, t)
		if err != nil {
			return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
		}

		tf.tfs[i] = tfer
	}

	return &tf, nil
}

type metaForEach struct {
	conf metaForEachConfig

	tf  Transformer
	tfs []Transformer
}

func (tf *metaForEach) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	var msgs []*message.Message
	var err error

	if msg.IsControl() {
		if len(tf.tfs) > 0 {
			msgs, err = Apply(ctx, tf.tfs, msg)
		} else {
			msgs, err = tf.tf.Transform(ctx, msg)
		}

		if err != nil {
			return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
		}

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
		if len(tf.tfs) > 0 {
			msgs, err = Apply(ctx, tf.tfs, tmpMsg)
		} else {
			msgs, err = tf.tf.Transform(ctx, tmpMsg)
		}

		if err != nil {
			return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
		}

		for _, m := range msgs {
			v := bytesToValue(m.Data())
			arr = append(arr, v.Value())
		}
	}

	if err := msg.SetValue(tf.conf.Object.TargetKey, arr); err != nil {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
	}

	return []*message.Message{msg}, nil
}

func (tf *metaForEach) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
