package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/internal/json"
)

/*
ZipOptions contain custom options settings for this processor.

Keys: location where elements from the input keys are written to; this creates JSON objects (optional)
*/
type ZipOptions struct {
	Keys []string `mapstructure:"keys"`
}

// Zip implements the Byter and Channeler interfaces and concatenates JSON arrays into an array of tuples or array of JSON objects. More information is available in the README.
type Zip struct {
	Condition condition.OperatorConfig `mapstructure:"condition"`
	Input     Inputs                   `mapstructure:"input"`
	Output    Output                   `mapstructure:"output"`
	Options   ZipOptions               `mapstructure:"options"`
}

// Channel processes a data channel of bytes with this processor. Conditions can be optionally applied on the channel data to enable processing.
func (p Zip) Channel(ctx context.Context, ch <-chan []byte) (<-chan []byte, error) {
	var array [][]byte

	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, err
	}

	for data := range ch {
		ok, err := op.Operate(data)
		if err != nil {
			return nil, err
		}

		if !ok {
			array = append(array, data)
			continue
		}

		processed, err := p.Byte(ctx, data)
		if err != nil {
			return nil, err
		}
		array = append(array, processed)
	}

	output := make(chan []byte, len(array))
	for _, x := range array {
		output <- x
	}
	close(output)
	return output, nil

}

// Byte processes a byte slice with this processor
func (p Zip) Byte(ctx context.Context, data []byte) ([]byte, error) {
	if len(p.Options.Keys) == 0 {
		// elements in the values array are stored at their
		// relative position inside the map to maintain order
		//
		// input.keys: ["foo","bar"], [123,456]
		// 	cache[0][]interface{}{"foo",123}
		// 	cache[1][]interface{}{"bar",456}
		cache := make(map[int][]interface{})
		for _, key := range p.Input.Keys {
			value := json.Get(data, key)
			for x, v := range value.Array() {
				cache[x] = append(cache[x], v.Value())
			}
		}

		var array []interface{}
		for _, v := range cache {
			array = append(array, v)
		}
		// [["foo",123],["bar",456]]
		return json.Set(data, p.Output.Key, array)
	}

	// elements in the values array are stored at their
	// 	relative position inside the map to maintain order
	//
	// input.keys: ["foo","bar"], [123,456]
	// options.keys: ["name","size"]
	// 	cache[0][]byte(`{"name":"foo","size":123}`)
	// 	cache[1][]byte(`{"name":"bar","size":456}`)
	cache := make(map[int][]byte)
	var err error
	for idx, key := range p.Input.Keys {
		value := json.Get(data, key)
		for x, v := range value.Array() {
			cache[x], err = json.Set(cache[x], p.Options.Keys[idx], v)
			if err != nil {
				return nil, err
			}
		}
	}

	// inserts pre-formatted JSON into an array based
	// 	on the length of the map
	// pre-formatted JSON requires use of SetRaw
	var tmp []byte
	for i, v := range cache {
		tmp, err = json.SetRaw(tmp, fmt.Sprintf("%d", i), v)
		if err != nil {
			return nil, err
		}
	}
	// [{"name":"foo","size":123},{"name":"bar","size":456}]
	return json.SetRaw(data, p.Output.Key, tmp)
}
