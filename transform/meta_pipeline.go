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

// errMetaPipelineArrayInput is returned when the transform is configured to process
// an object, but the input is an array. Array values are not supported by this transform,
// instead the input should be run through the metaForEach transform (which can encapsulate
// the pipeline transform).
var errMetaPipelineArrayInput = fmt.Errorf("input is an array")

type metaPipelineConfig struct {
	// Transforms that are applied in series to the data.
	Transforms []config.Config `json:"transforms"`

	ID     string         `json:"id"`
	Object iconfig.Object `json:"object"`
}

func (c *metaPipelineConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *metaPipelineConfig) Validate() error {
	if c.Object.SourceKey == "" && c.Object.TargetKey != "" {
		return fmt.Errorf("object_source_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.Object.SourceKey != "" && c.Object.TargetKey == "" {
		return fmt.Errorf("object_target_key: %v", errors.ErrMissingRequiredOption)
	}

	if len(c.Transforms) == 0 {
		return fmt.Errorf("transforms: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

// Deprecated: newMetaPipeline exists for backwards compatibility and will be
// removed in a future release. Use the Transforms fields on other meta transforms
// instead.
func newMetaPipeline(ctx context.Context, cfg config.Config) (*metaPipeline, error) {
	conf := metaPipelineConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform meta_pipeline: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "meta_pipeline"
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	tf := metaPipeline{
		conf:     conf,
		isObject: conf.Object.SourceKey != "" && conf.Object.TargetKey != "",
	}

	var tform []Transformer
	for _, c := range conf.Transforms {
		t, err := New(ctx, c)
		if err != nil {
			return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
		}

		tform = append(tform, t)
	}
	tf.tf = tform

	return &tf, nil
}

type metaPipeline struct {
	conf     metaPipelineConfig
	isObject bool

	tf []Transformer
}

func (tf *metaPipeline) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		msgs, err := Apply(ctx, tf.tf, msg)
		if err != nil {
			return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
		}

		return msgs, nil
	}

	if !tf.isObject {
		msgs, err := Apply(ctx, tf.tf, msg)
		if err != nil {
			return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
		}

		return msgs, nil
	}

	value := msg.GetValue(tf.conf.Object.SourceKey)
	if !value.Exists() {
		return []*message.Message{msg}, nil
	}

	if value.IsArray() {
		return nil, fmt.Errorf("transform %s: key %s: %v", tf.conf.ID, tf.conf.Object.SourceKey, errMetaPipelineArrayInput)
	}

	res, err := Apply(ctx, tf.tf, message.New().SetData(value.Bytes()))
	if err != nil {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
	}

	var output []*message.Message
	for _, msg := range res {
		if err := msg.SetValue(tf.conf.Object.TargetKey, msg.Data()); err != nil {
			return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
		}

		output = append(output, msg)
	}

	return output, nil
}

func (tf *metaPipeline) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
