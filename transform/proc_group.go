package transform

import (
	"context"
	gojson "encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	_config "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/json"
	mess "github.com/brexhq/substation/message"
)

type procGroupConfig struct {
	// Key retrieves a value from an object for processing.
	//
	// This is optional for transforms that support processing non-object data.
	Key string `json:"key"`
	// SetKey inserts a processed value into an object.
	//
	// This is optional for transforms that support processing non-object data.
	SetKey string `json:"set_key"`
	// Keys determines where processed values are set in newly created objects.
	//
	// This is optional and defaults to creating an array of tuples instead
	// of an array of objects.
	Keys []string `json:"keys"`
}

type procGroup struct {
	conf procGroupConfig
}

func newProcGroup(_ context.Context, cfg config.Config) (*procGroup, error) {
	conf := procGroupConfig{}
	if err := _config.Decode(cfg.Settings, &conf); err != nil {
		return nil, err
	}

	// Validate required options.
	if conf.Key == "" || conf.SetKey == "" {
		return nil, fmt.Errorf("transform: proc_group: key %s set_key %s: %v", conf.Key, conf.SetKey, errInvalidDataPattern)
	}

	proc := procGroup{
		conf: conf,
	}

	return &proc, nil
}

func (proc *procGroup) String() string {
	b, _ := gojson.Marshal(proc.conf)
	return string(b)
}

func (*procGroup) Close(context.Context) error {
	return nil
}

func (proc *procGroup) Transform(ctx context.Context, message *mess.Message) ([]*mess.Message, error) {
	// Skip control messages.
	if message.IsControl() {
		return []*mess.Message{message}, nil
	}

	if len(proc.conf.Keys) == 0 {
		// Elements in the values array are stored at their
		// relative position inside the map to maintain order.
		//
		// input.key: [["foo","bar"],[123,456]]
		// 	cache[0][]interface{}{"foo",123}
		// 	cache[1][]interface{}{"bar",456}
		cache := make(map[int][]interface{})
		result := message.Get(proc.conf.Key)
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
		if err := message.Set(proc.conf.SetKey, value); err != nil {
			return nil, fmt.Errorf("transform: proc_group: %v", err)
		}

		return []*mess.Message{message}, nil
	}

	// Elements in the values array are stored at their
	// relative position inside the map to maintain order
	//
	// input.key: [["foo","bar"],[123,456]]
	// options.keys: ["name","size"]
	// 	cache[0][]byte(`{"name":"foo","size":123}`)
	// 	cache[1][]byte(`{"name":"bar","size":456}`)
	cache := make(map[int][]byte)

	var err error
	result := message.Get(proc.conf.Key)
	for i, res := range result.Array() {
		for j, r := range res.Array() {
			cache[j], err = json.Set(cache[j], proc.conf.Keys[i], r)
			if err != nil {
				return nil, fmt.Errorf("transform: proc_group: %v", err)
			}
		}
	}

	// Inserts pre-formatted JSON into an array based
	// on the length of the mat.
	var value []byte
	for i := 0; i < len(cache); i++ {
		value, err = json.Set(value, fmt.Sprintf("%d", i), cache[i])
		if err != nil {
			return nil, fmt.Errorf("transform: proc_group: %v", err)
		}
	}

	if err := message.Set(proc.conf.SetKey, value); err != nil {
		return nil, fmt.Errorf("transform: proc_group: %v", err)
	}

	return []*mess.Message{message}, nil
}
