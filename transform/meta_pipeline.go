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
	Object iconfig.Object `json:"object"`

	// Transforms applied in series to the data.
	Transforms []config.Config `json:"transforms"`
}

func (c *metaPipelineConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *metaPipelineConfig) Validate() error {
	if c.Object.Key == "" && c.Object.SetKey != "" {
		return fmt.Errorf("object_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.Object.Key != "" && c.Object.SetKey == "" {
		return fmt.Errorf("object_set_key: %v", errors.ErrMissingRequiredOption)
	}

	if len(c.Transforms) == 0 {
		return fmt.Errorf("transforms: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

func newMetaPipeline(ctx context.Context, cfg config.Config) (*metaPipeline, error) {
	conf := metaPipelineConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: meta_pipeline: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: meta_pipeline: %v", err)
	}

	tf := metaPipeline{
		conf:     conf,
		isObject: conf.Object.Key != "" && conf.Object.SetKey != "",
	}

	var tform []Transformer
	for _, c := range conf.Transforms {
		t, err := New(ctx, c)
		if err != nil {
			return nil, fmt.Errorf("transform: meta_pipeline: transform %+v: %v", c, err)
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
		return []*message.Message{msg}, nil
	}

	if !tf.isObject {
		msgs, err := Apply(ctx, tf.tf, msg)
		if err != nil {
			return nil, fmt.Errorf("transform: meta_pipeline: %v", err)
		}

		return msgs, nil
	}

	value := msg.GetValue(tf.conf.Object.Key)
	if !value.Exists() {
		return []*message.Message{msg}, nil
	}

	if value.IsArray() {
		return nil, fmt.Errorf("transform: meta_pipeline: key %s: %v", tf.conf.Object.Key, errMetaPipelineArrayInput)
	}

	tmpMsg := message.New().SetData(value.Bytes())
	msgs, err := Apply(ctx, tf.tf, tmpMsg)
	if err != nil {
		return nil, fmt.Errorf("transform: meta_pipeline: %v", err)
	}

	var output []*message.Message
	for _, msg := range msgs {
		if err := msg.SetValue(tf.conf.Object.SetKey, msg.Data()); err != nil {
			return nil, fmt.Errorf("transform: meta_pipeline: %v", err)
		}

		output = append(output, msg)
	}

	return output, nil
}

func (tf *metaPipeline) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
