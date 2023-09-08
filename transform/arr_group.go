package transform

import (
	"context"
	gojson "encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/message"
	"github.com/tidwall/sjson"
)

type arrGroupConfig struct {
	Object iconfig.Object `json:"object"`

	// Keys determines where values are set in newly created objects.
	//
	// This is optional and defaults to creating an array of tuples instead
	// of an array of objects.
	Keys []string `json:"keys"`
}

func (c *arrGroupConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *arrGroupConfig) Validate() error {
	if c.Object.Key == "" && c.Object.SetKey != "" {
		return fmt.Errorf("object_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.Object.Key != "" && c.Object.SetKey == "" {
		return fmt.Errorf("object_set_key: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

type arrGroup struct {
	conf arrGroupConfig
}

func newArrGroup(_ context.Context, cfg config.Config) (*arrGroup, error) {
	conf := arrGroupConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: new_array_group: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: new_array_group: %v", err)
	}

	tf := arrGroup{
		conf: conf,
	}

	return &tf, nil
}

func (tf *arrGroup) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	// Skip interrupt messages.
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	if len(tf.conf.Keys) == 0 {
		// Elements in the values array are stored at their
		// relative position inside the map to maintain order.
		//
		// input.key: [["a","b"],[1,2]]
		// 	cache[0][]interface{}{"a",1}
		// 	cache[1][]interface{}{"b",2}
		cache := make(map[int][]interface{})
		value := msg.GetValue(tf.conf.Object.Key)
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
	cache := make(map[int][]byte)

	var err error
	value := msg.GetValue(tf.conf.Object.Key)
	for i, val := range value.Array() {
		for j, v := range val.Array() {
			cache[j], err = sjson.SetBytes(cache[j], tf.conf.Keys[i], v.Value())
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

func (tf *arrGroup) String() string {
	b, _ := gojson.Marshal(tf.conf)
	return string(b)
}

func (*arrGroup) Close(context.Context) error {
	return nil
}
