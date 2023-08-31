package transform

import (
	"context"
	gojson "encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/json"
	"github.com/brexhq/substation/message"
)

type modGroupConfig struct {
	Object configObject `json:"object"`

	// Keys determines where values are set in newly created objects.
	//
	// This is optional and defaults to creating an array of tuples instead
	// of an array of objects.
	Keys []string `json:"keys"`
}

type modGroup struct {
	conf modGroupConfig
}

func newModGroup(_ context.Context, cfg config.Config) (*modGroup, error) {
	conf := modGroupConfig{}
	if err := iconfig.Decode(cfg.Settings, &conf); err != nil {
		return nil, fmt.Errorf("transform: new_mod_group: %v", err)
	}

	// Validate required options.
	if conf.Object.Key == "" && conf.Object.SetKey != "" {
		return nil, fmt.Errorf("transform: new_mod_group: object_key: %v", errors.ErrMissingRequiredOption)
	}

	if conf.Object.Key != "" && conf.Object.SetKey == "" {
		return nil, fmt.Errorf("transform: new_mod_group: object_set_key: %v", errors.ErrMissingRequiredOption)
	}

	tf := modGroup{
		conf: conf,
	}

	return &tf, nil
}

func (tf *modGroup) String() string {
	b, _ := gojson.Marshal(tf.conf)
	return string(b)
}

func (*modGroup) Close(context.Context) error {
	return nil
}

func (tf *modGroup) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
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
		result := msg.GetObject(tf.conf.Object.Key)
		for _, res := range result.Array() {
			for i, r := range res.Array() {
				cache[i] = append(cache[i], r.Value())
			}
		}

		var value []interface{}
		for i := 0; i < len(cache); i++ {
			value = append(value, cache[i])
		}

		// [["foo",123],["bar",456]]
		if err := msg.SetObject(tf.conf.Object.SetKey, value); err != nil {
			return nil, fmt.Errorf("transform: mod_group: %v", err)
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
	result := msg.GetObject(tf.conf.Object.Key)
	for i, res := range result.Array() {
		for j, r := range res.Array() {
			cache[j], err = json.Set(cache[j], tf.conf.Keys[i], r)
			if err != nil {
				return nil, fmt.Errorf("transform: mod_group: %v", err)
			}
		}
	}

	// Inserts pre-formatted JSON into an array based
	// on the length of the result.
	var value []byte
	for i := 0; i < len(cache); i++ {
		value, err = json.Set(value, fmt.Sprintf("%d", i), cache[i])
		if err != nil {
			return nil, fmt.Errorf("transform: mod_group: %v", err)
		}
	}

	if err := msg.SetObject(tf.conf.Object.SetKey, value); err != nil {
		return nil, fmt.Errorf("transform: mod_group: %v", err)
	}

	return []*message.Message{msg}, nil
}
