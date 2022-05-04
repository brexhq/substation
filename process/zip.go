package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/json"
)

// ZipInvalidSettings is returned when the Zip processor is configured with invalid Input and Output settings.
const ZipInvalidSettings = errors.Error("ZipInvalidSettings")

/*
ZipOptions contains custom options for the Zip processor:
	Keys:
		where values from Inputs.Keys are written to, creating new JSON objects
*/
type ZipOptions struct {
	Keys []string `mapstructure:"keys"`
}

/*
Zip processes data by grouping JSON arrays into an array of tuples or array of JSON objects. The processor supports these patterns:
	json array:
		{"names":["foo","bar"],"sizes":[111,222]} >>> {"names":["foo","bar"],"sizes":[111,222],"group":[["foo",111],["bar",222]]}
		{"names":["foo","bar"],"sizes":[111,222]} >>> {"names":["foo","bar"],"sizes":[111,222],"group":[{"name":foo","size":111},{"name":"bar","size":222}]}

The processor uses this Jsonnet configuration:
	{
		type: 'group',
		settings: {
			// if the values are ["foo","bar"] and [123,456], then this returns [["foo",123],["bar",456]]
			input: {
				keys: ['names','sizes'],
			},
			output: {
				key: 'group',
			}
		},
	}
*/
type Zip struct {
	Condition condition.OperatorConfig `mapstructure:"condition"`
	Input     Inputs                   `mapstructure:"input"`
	Output    Output                   `mapstructure:"output"`
	Options   ZipOptions               `mapstructure:"options"`
}

// Channel processes a data channel of byte slices with the Zip processor. Conditions are optionally applied on the channel data to enable processing.
func (p Zip) Channel(ctx context.Context, ch <-chan []byte) (<-chan []byte, error) {
	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, err
	}

	var array [][]byte
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

// Byte processes a byte slice with the Zip processor.
func (p Zip) Byte(ctx context.Context, data []byte) ([]byte, error) {
	// only supports json arrays, so error early if there are no keys
	if len(p.Input.Keys) == 0 && p.Output.Key == "" {
		return nil, ZipInvalidSettings
	}

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
