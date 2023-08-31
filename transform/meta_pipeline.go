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

// errMetaPipelineArrayInput is returned when the transform is configured to process
// an object, but the input is an array. Array values are not supported by this transform,
// instead the input should be run through the metaForEach transform (which can encapsulate
// the pipeline transform).
var errMetaPipelineArrayInput = fmt.Errorf("input is an array")

type metaPipelineConfig struct {
	Object configObject `json:"object"`

	// Transforms applied in series to the data.
	Transforms []config.Config `json:"transforms"`
}

type metaPipeline struct {
	conf     metaPipelineConfig
	isObject bool

	tf []Transformer
}

func newMetaPipeline(ctx context.Context, cfg config.Config) (*metaPipeline, error) {
	conf := metaPipelineConfig{}
	if err := iconfig.Decode(cfg.Settings, &conf); err != nil {
		return nil, fmt.Errorf("transform: new_meta_pipeline: %v", err)
	}

	// Validate required options.
	if conf.Object.Key == "" && conf.Object.SetKey != "" {
		return nil, fmt.Errorf("transform: new_meta_pipeline: object_key: %v", errors.ErrMissingRequiredOption)
	}

	if conf.Object.Key != "" && conf.Object.SetKey == "" {
		return nil, fmt.Errorf("transform: new_meta_pipeline: object_set_key: %v", errors.ErrMissingRequiredOption)
	}

	if len(conf.Transforms) == 0 {
		return nil, fmt.Errorf("transform: new_meta_pipeline: transforms: %v", errors.ErrMissingRequiredOption)
	}

	meta := metaPipeline{
		conf:     conf,
		isObject: conf.Object.Key != "" && conf.Object.SetKey != "",
	}

	var tf []Transformer
	for _, c := range conf.Transforms {
		t, err := New(ctx, c)
		if err != nil {
			return nil, fmt.Errorf("transform: new_meta_pipeline: transform %+v: %v", c, err)
		}

		tf = append(tf, t)
	}
	meta.tf = tf

	return &meta, nil
}

func (meta *metaPipeline) String() string {
	b, _ := gojson.Marshal(meta.conf)
	return string(b)
}

func (*metaPipeline) Close(context.Context) error {
	return nil
}

func (meta *metaPipeline) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	// Skip interrupt messages.
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	if !meta.isObject {
		msgs, err := Apply(ctx, meta.tf, msg)
		if err != nil {
			return nil, fmt.Errorf("transform: meta_pipeline: %v", err)
		}

		return msgs, nil
	}

	result := msg.GetObject(meta.conf.Object.Key)
	if result.IsArray() {
		return nil, fmt.Errorf("transform: meta_pipeline: key %s: %v", meta.conf.Object.Key, errMetaPipelineArrayInput)
	}

	newMsg := message.New().SetData(result.Bytes())
	msgs, err := Apply(ctx, meta.tf, newMsg)
	if err != nil {
		return nil, fmt.Errorf("transform: meta_pipeline: %v", err)
	}

	var newMsgs []*message.Message
	for _, msg := range msgs {
		if err := msg.SetObject(meta.conf.Object.SetKey, msg.Data()); err != nil {
			return nil, fmt.Errorf("transform: meta_pipeline: %v", err)
		}

		newMsgs = append(newMsgs, msg)
	}

	return newMsgs, nil
}
