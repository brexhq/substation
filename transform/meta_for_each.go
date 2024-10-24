package transform

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"

	iconfig "github.com/brexhq/substation/v2/internal/config"
)

type metaForEachConfig struct {
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
		return fmt.Errorf("object_source_key: %v", iconfig.ErrMissingRequiredOption)
	}

	if c.Object.TargetKey == "" {
		return fmt.Errorf("object_target_key: %v", iconfig.ErrMissingRequiredOption)
	}

	if len(c.Transforms) == 0 {
		return fmt.Errorf("transforms: %v", iconfig.ErrMissingRequiredOption)
	}

	for _, t := range c.Transforms {
		if t.Type == "" {
			return fmt.Errorf("transform: %v", iconfig.ErrMissingRequiredOption)
		}
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

	tfs []Transformer
}

func (tf *metaForEach) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.HasFlag(message.IsControl) {
		msgs, err := Apply(ctx, tf.tfs, msg)
		if err != nil {
			return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
		}

		return msgs, nil
	}

	value := msg.GetValue(tf.conf.Object.SourceKey)
	if skipMessage(msg, value) {
		return []*message.Message{msg}, nil
	}

	if !value.IsArray() {
		return []*message.Message{msg}, nil
	}

	var arr []interface{}
	for _, res := range value.Array() {
		m := message.New().SetData(res.Bytes()).SetMetadata(msg.Metadata())
		msgs, err := Apply(ctx, tf.tfs, m)
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
