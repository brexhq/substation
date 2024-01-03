package transform

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/message"
	"github.com/tidwall/sjson"
)

type arrayToObjectConfig struct {
	Object     iconfig.Object `json:"object"`
	ObjectKeys []string       `json:"object_keys"`
}

func (c *arrayToObjectConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *arrayToObjectConfig) Validate() error {
	if c.Object.SourceKey == "" && c.Object.TargetKey != "" {
		return fmt.Errorf("object_source_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.Object.SourceKey != "" && c.Object.TargetKey == "" {
		return fmt.Errorf("object_target_key: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

func newArrayToObject(_ context.Context, cfg config.Config) (*arrayToObject, error) {
	conf := arrayToObjectConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: array_to_object: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: array_to_object: %v", err)
	}

	tf := arrayToObject{
		conf:      conf,
		hasObjSrc: conf.Object.SourceKey != "",
		hasObjDst: conf.Object.TargetKey != "",
	}

	return &tf, nil
}

type arrayToObject struct {
	conf arrayToObjectConfig

	hasObjSrc bool
	hasObjDst bool
}

func (tf *arrayToObject) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	var value message.Value
	if tf.hasObjSrc {
		value = msg.GetValue(tf.conf.Object.SourceKey)
	} else {
		value = bytesToValue(msg.Data())
	}

	if !value.Exists() || !value.IsArray() {
		return []*message.Message{msg}, nil
	}

	cache := make(map[int][]message.Value)

	if len(tf.conf.ObjectKeys) > 0 {
		for i, key := range tf.conf.ObjectKeys {
			v := bytesToValue([]byte(key))
			cache[i] = append(cache[i], v)
		}
	}

	for _, val := range value.Array() {
		for i, v := range val.Array() {
			cache[i] = append(cache[i], v)
		}
	}

	var b []byte
	var err error
	for idx := 0; idx < len(cache); idx++ {
		switch len(cache[idx]) {
		case 0, 1:
			continue
		case 2:
			b, err = sjson.SetBytes(b, cache[idx][0].String(), cache[idx][1].Value())
			if err != nil {
				return nil, fmt.Errorf("transform: array_to_object: %v", err)
			}
		default:
			var vals []interface{}

			for i, v := range cache[idx] {
				if i == 0 {
					continue
				}

				vals = append(vals, v.Value())
			}

			b, err = sjson.SetBytes(b, cache[idx][0].String(), vals)
			if err != nil {
				return nil, fmt.Errorf("transform: array_to_object: %v", err)
			}
		}
	}

	if tf.hasObjDst {
		if err := msg.SetValue(tf.conf.Object.TargetKey, b); err != nil {
			return nil, fmt.Errorf("transform: array_to_object: %v", err)
		}

		return []*message.Message{msg}, nil
	}

	msg.SetData(b)
	return []*message.Message{msg}, nil
}

func (tf *arrayToObject) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
