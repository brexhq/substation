package transform

import (
	"context"
	gojson "encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	_config "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/json"
	mess "github.com/brexhq/substation/message"
)

type metaForEachConfig struct {
	// Key retrieves a value from an object for processing.
	//
	// This is optional for transforms that support processing non-object data.
	Key string `json:"key"`
	// SetKey inserts a processed value into an object.
	//
	// This is optional for transforms that support processing non-object data.
	SetKey string `json:"set_key"`
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
	if err := _config.Decode(cfg.Settings, &conf); err != nil {
		return nil, err
	}

	// Validate required options.
	if conf.Key == "" || conf.SetKey == "" {
		return nil, fmt.Errorf("transform: meta_for_each: key %s set_key %s: %v", conf.Key, conf.SetKey, errInvalidDataPattern)
	}

	if conf.Transform.Type == "" {
		return nil, fmt.Errorf("transform: meta_for_each: type: %v", errors.ErrMissingRequiredOption)
	}

	tfConf, err := gojson.Marshal(conf.Transform)
	if err != nil {
		return nil, err
	}

	inputKey := conf.Transform.Type
	if innerKey, ok := conf.Transform.Settings["key"].(string); ok && innerKey != "" {
		inputKey = conf.Transform.Type + "." + innerKey
	}
	tfConf, _ = json.Set(tfConf, "settings.key", inputKey)

	outputKey := conf.Transform.Type
	if innerKey, ok := conf.Transform.Settings["set_key"].(string); ok && innerKey != "" {
		outputKey = conf.Transform.Type + "." + innerKey
	}
	tfConf, _ = json.Set(tfConf, "settings.set_key", outputKey)

	meta := metaForEach{
		conf: conf,
	}

	if err := gojson.Unmarshal(tfConf, &meta.tfCfg); err != nil {
		return nil, err
	}

	tf, err := New(ctx, meta.tfCfg)
	if err != nil {
		return nil, fmt.Errorf("process: for_each: %v", err)
	}
	meta.tf = tf

	return &meta, nil
}

func (meta *metaForEach) String() string {
	b, _ := gojson.Marshal(meta.conf)
	return string(b)
}

func (*metaForEach) Close(context.Context) error {
	return nil
}

func (meta *metaForEach) Transform(ctx context.Context, message *mess.Message) ([]*mess.Message, error) {
	if message.IsControl() {
		return []*mess.Message{message}, nil
	}

	result := message.Get(meta.conf.Key)
	if !result.IsArray() {
		return []*mess.Message{message}, nil
	}

	var arr []interface{}
	for _, res := range result.Array() {
		tmpMsg, err := mess.New()
		if err != nil {
			return nil, fmt.Errorf("transform: meta_for_each: %v", err)
		}

		if err := tmpMsg.Set(meta.tfCfg.Type, res); err != nil {
			return nil, fmt.Errorf("transform: meta_for_each: %v", err)
		}

		tfMsgs, err := meta.tf.Transform(ctx, tmpMsg)
		if err != nil {
			return nil, fmt.Errorf("transform: meta_for_each: %v", err)
		}

		for _, m := range tfMsgs {
			v := m.Get(meta.tfCfg.Type)
			arr = append(arr, v.Value())
		}
	}

	if err := message.Set(meta.conf.SetKey, arr); err != nil {
		return nil, fmt.Errorf("transform: meta_for_each: %v", err)
	}

	return []*mess.Message{message}, nil
}
