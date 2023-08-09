package transform

import (
	"context"
	gojson "encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	_config "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	mess "github.com/brexhq/substation/message"
)

// errMetaPipelineArrayInput is returned when the transform is configured to process
// an object, but the input is an array. Array values are not supported by this transform,
// instead the input should be run through the metaForEach transform (which can encapsulate
// the pipeline transform).
var errMetaPipelineArrayInput = fmt.Errorf("input is an array")

type metaPipelineConfig struct {
	// Key retrieves a value from an object for processing.
	//
	// This is optional for transforms that support processing non-object data.
	Key string `json:"key"`
	// SetKey inserts a processed value into an object.
	//
	// This is optional for transforms that support processing non-object data.
	SetKey string `json:"set_key"`
	// Transforms applied in series to the data.
	Transforms []config.Config `json:"transforms"`
}

type metaPipeline struct {
	conf     metaPipelineConfig
	isObject bool

	transforms []Transformer
}

func newMetaPipeline(ctx context.Context, cfg config.Config) (*metaPipeline, error) {
	conf := metaPipelineConfig{}
	if err := _config.Decode(cfg.Settings, &conf); err != nil {
		return nil, err
	}

	// Validate required options.
	if (conf.Key == "" && conf.SetKey != "") || (conf.Key != "" && conf.SetKey == "") {
		return nil, fmt.Errorf("transform: meta_pipeline: key %s set_key %s: %v", conf.Key, conf.SetKey, errInvalidDataPattern)
	}

	if len(conf.Transforms) == 0 {
		return nil, fmt.Errorf("transform: meta_pipeline: transforms: %v", errors.ErrMissingRequiredOption)
	}

	meta := metaPipeline{
		conf:     conf,
		isObject: conf.Key != "" && conf.SetKey != "",
	}

	var tforms []Transformer
	for _, c := range conf.Transforms {
		t, err := New(ctx, c)
		if err != nil {
			return nil, fmt.Errorf("transform: meta_pipeline: transform %+v: %v", c, err)
		}

		tforms = append(tforms, t)
	}
	meta.transforms = tforms

	return &meta, nil
}

func (t *metaPipeline) String() string {
	b, _ := gojson.Marshal(t.conf)
	return string(b)
}

func (*metaPipeline) Close(context.Context) error {
	return nil
}

func (t *metaPipeline) Transform(ctx context.Context, messages ...*mess.Message) ([]*mess.Message, error) {
	var output []*mess.Message

	for _, message := range messages {
		// Skip control messages.
		if message.IsControl() {
			output = append(output, message)
			continue
		}

		switch t.isObject {
		case true:
			result := message.Get(t.conf.Key)
			if result.IsArray() {
				return nil, fmt.Errorf("transform: meta_pipeline: key %s: %v", t.conf.Key, errMetaPipelineArrayInput)
			}

			newMessage, err := mess.New(
				mess.SetData([]byte(result.String())),
			)
			if err != nil {
				return nil, fmt.Errorf("transform: meta_pipeline: %v", err)
			}

			nMessage, err := Apply(ctx, t.transforms, newMessage)
			if err != nil {
				return nil, fmt.Errorf("transform: meta_pipeline: %v", err)
			}

			var tmpMessages []*mess.Message
			for _, message := range nMessage {
				msg, err := mess.New(
					mess.SetData(message.Data()),
					mess.SetMetadata(message.Metadata()),
				)
				if err != nil {
					return nil, fmt.Errorf("transform: meta_pipeline: %v", err)
				}

				if err := msg.Set(t.conf.SetKey, message.Data()); err != nil {
					return nil, fmt.Errorf("transform: meta_pipeline: %v", err)
				}

				tmpMessages = append(tmpMessages, msg)
			}

			output = append(output, tmpMessages...)
		case false:
			msg, err := Apply(ctx, t.transforms, message)
			if err != nil {
				return nil, fmt.Errorf("transform: meta_pipeline: %v", err)
			}

			output = append(output, msg...)
		}
	}

	return output, nil
}
