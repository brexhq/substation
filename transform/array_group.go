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

type arrayGroupConfig struct {
	Object iconfig.Object `json:"object"`

	// GroupKeys determines where values are set in newly created objects.
	//
	// This is optional and defaults to creating an array of tuples instead
	// of an array of objects.
	GroupKeys []string `json:"group_keys"`
}

func (c *arrayGroupConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *arrayGroupConfig) Validate() error {
	if c.Object.Key == "" {
		return fmt.Errorf("object_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.Object.SetKey == "" {
		return fmt.Errorf("object_set_key: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

func newArrayGroup(_ context.Context, cfg config.Config) (*arrayGroup, error) {
	conf := arrayGroupConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: array_group: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: array_group: %v", err)
	}

	tf := arrayGroup{
		conf: conf,
	}

	return &tf, nil
}

type arrayGroup struct {
	conf arrayGroupConfig
}

func (tf *arrayGroup) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	if len(tf.conf.GroupKeys) == 0 {
		// Elements in the values array are stored at their
		// relative position inside the map to maintain order.
		//
		// input.key: [["a","b"],[1,2]]
		// 	cache[0][]interface{}{"a",1}
		// 	cache[1][]interface{}{"b",2}
		value := msg.GetValue(tf.conf.Object.Key)
		if !value.Exists() {
			return []*message.Message{msg}, nil
		}

		if !value.IsArray() {
			return []*message.Message{msg}, nil
		}

		cache := make(map[int][]interface{})
		for _, val := range value.Array() {
			for i, v := range val.Array() {
				cache[i] = append(cache[i], v.Value())
			}
		}

		var b []interface{}
		for i := 0; i < len(cache); i++ {
			b = append(b, cache[i])
		}

		// [["foo",123],["bar",456]]
		if err := msg.SetValue(tf.conf.Object.SetKey, b); err != nil {
			return nil, fmt.Errorf("transform: array_group: %v", err)
		}

		return []*message.Message{msg}, nil
	}

	// Elements in the values array are stored at their
	// relative position inside the map to maintain order.
	//
	// input.key: [["a","b"],[1,2]]
	// options.keys: ["str","int"]
	// 	cache[0][]byte(`{"str":"a","int":1}`)
	// 	cache[1][]byte(`{"str":"b","int":2}`)
	value := msg.GetValue(tf.conf.Object.Key)
	if !value.Exists() {
		return []*message.Message{msg}, nil
	}

	if !value.IsArray() {
		return []*message.Message{msg}, nil
	}

	var err error
	cache := make(map[int][]byte)

	for i, val := range value.Array() {
		for j, v := range val.Array() {
			cache[j], err = sjson.SetBytes(cache[j], tf.conf.GroupKeys[i], v.Value())
			if err != nil {
				return nil, fmt.Errorf("transform: array_group: %v", err)
			}
		}
	}

	// Inserts pre-formatted JSON into an array based
	// on the length of the result. SetRawBytes is used
	// to avoid re-encoding the JSON.
	var b []byte
	for i := 0; i < len(cache); i++ {
		b, err = sjson.SetRawBytes(b, fmt.Sprintf("%d", i), cache[i])
		if err != nil {
			return nil, fmt.Errorf("transform: array_group: %v", err)
		}
	}

	if err := msg.SetValue(tf.conf.Object.SetKey, b); err != nil {
		return nil, fmt.Errorf("transform: array_group: %v", err)
	}

	return []*message.Message{msg}, nil
}

func (tf *arrayGroup) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
