package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/internal/json"
)

/*
GroupOptions contains custom options for the Group processor:
	Keys (optional):
		where values from Inputs.Keys are written to, creating new JSON objects
*/
type GroupOptions struct {
	Keys []string `json:"keys"`
}

/*
Group processes data by grouping JSON arrays into an array of tuples or array of JSON objects. The processor supports these patterns:
	json array:
		{"group":[["foo","bar"],[111,222]]} >>> {"group":[["foo",111],["bar",222]]}
		{"group":[["foo","bar"],[111,222]]} >>> {"group":[{"name":foo","size":111},{"name":"bar","size":222}]}

The processor uses this Jsonnet configuration:
	{
		type: 'group',
		settings: {
			input_key: 'group',
			output_key: 'group',
			}
		},
	}
*/
type Group struct {
	Options   GroupOptions             `json:"options"`
	Condition condition.OperatorConfig `json:"condition"`
	InputKey  string                   `json:"input_key"`
	OutputKey string                   `json:"output_key"`
}

// Slice processes a slice of bytes with the Group processor. Conditions are optionally applied on the bytes to enable processing.
func (p Group) Slice(ctx context.Context, s [][]byte) ([][]byte, error) {
	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, fmt.Errorf("slicer settings %v: %v", p, err)
	}

	slice := NewSlice(&s)
	for _, data := range s {
		ok, err := op.Operate(data)
		if err != nil {
			return nil, fmt.Errorf("slicer settings %v: %v", p, err)
		}

		if !ok {
			slice = append(slice, data)
			continue
		}

		processed, err := p.Byte(ctx, data)
		if err != nil {
			return nil, fmt.Errorf("slicer: %v", err)
		}
		slice = append(slice, processed)
	}

	return slice, nil
}

// Byte processes bytes with the Group processor.
func (p Group) Byte(ctx context.Context, data []byte) ([]byte, error) {
	// only supports JSON arrays, error early if there are no keys
	if p.InputKey == "" && p.OutputKey == "" {
		return nil, fmt.Errorf("byter settings %+v: %v", p, ProcessorInvalidSettings)
	}

	if len(p.Options.Keys) == 0 {
		// elements in the values array are stored at their
		// relative position inside the map to maintain order
		//
		// input.key: [["foo","bar"],[123,456]]
		// 	cache[0][]interface{}{"foo",123}
		// 	cache[1][]interface{}{"bar",456}
		cache := make(map[int][]interface{})
		value := json.Get(data, p.InputKey)
		for _, v := range value.Array() {
			for x, v1 := range v.Array() {
				cache[x] = append(cache[x], v1.Value())
			}
		}

		var array []interface{}
		for i := 0; i < len(cache); i++ {
			array = append(array, cache[i])
		}

		// [["foo",123],["bar",456]]
		return json.Set(data, p.OutputKey, array)
	}

	// elements in the values array are stored at their
	// 	relative position inside the map to maintain order
	//
	// input.key: [["foo","bar"],[123,456]]
	// options.keys: ["name","size"]
	// 	cache[0][]byte(`{"name":"foo","size":123}`)
	// 	cache[1][]byte(`{"name":"bar","size":456}`)
	cache := make(map[int][]byte)
	var err error
	value := json.Get(data, p.InputKey)
	for x, v := range value.Array() {
		for x1, v1 := range v.Array() {
			cache[x1], err = json.Set(cache[x1], p.Options.Keys[x], v1)
			if err != nil {
				return nil, fmt.Errorf("byter settings %+v: %v", p, err)
			}
		}
	}

	// inserts pre-formatted JSON into an array based
	// on the length of the map
	var tmp []byte
	for i := 0; i < len(cache); i++ {
		tmp, err = json.Set(tmp, fmt.Sprintf("%d", i), cache[i])
		if err != nil {
			return nil, fmt.Errorf("byter settings %+v: %v", p, err)
		}
	}

	// JSON arrays must be set using SetRaw to preserve structure
	//
	// [{"name":"foo","size":123},{"name":"bar","size":456}]
	return json.SetRaw(data, p.OutputKey, tmp)
}
