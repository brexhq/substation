package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/json"
)

/*
Group processes data by grouping JSON arrays into an array of tuples or array of JSON objects. The processor supports these patterns:
	JSON array:
		{"group":[["foo","bar"],[111,222]]} >>> {"group":[["foo",111],["bar",222]]}
		{"group":[["foo","bar"],[111,222]]} >>> {"group":[{"name":foo","size":111},{"name":"bar","size":222}]}

When loaded with a factory, the processor uses this JSON configuration:
	{
		"type": "group",
		"settings": {
			"input_key": "group",
			"output_key": "group"
		}
	}
*/
type Group struct {
	Options   GroupOptions     `json:"options"`
	Condition condition.Config `json:"condition"`
	InputKey  string           `json:"input_key"`
	OutputKey string           `json:"output_key"`
}

/*
GroupOptions contains custom options for the Group processor:
	Keys (optional):
		where values from InputKey are written to, creating new JSON objects
*/
type GroupOptions struct {
	Keys []string `json:"keys"`
}

// ApplyBatch processes a slice of encapsulated data with the Group processor. Conditions are optionally applied to the data to enable processing.
func (p Group) ApplyBatch(ctx context.Context, caps []config.Capsule) ([]config.Capsule, error) {
	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, fmt.Errorf("applybatch settings %+v: %w", p, err)
	}

	caps, err = conditionallyApplyBatch(ctx, caps, op, p)
	if err != nil {
		return nil, fmt.Errorf("applybatch settings %+v: %w", p, err)
	}

	return caps, nil
}

// Apply processes encapsulated data with the Group processor.
func (p Group) Apply(ctx context.Context, cap config.Capsule) (config.Capsule, error) {
	// only supports JSON arrays, error early if there are no keys
	if p.InputKey == "" && p.OutputKey == "" {
		return cap, fmt.Errorf("applicator settings %+v: %w", p, ProcessorInvalidSettings)
	}

	if len(p.Options.Keys) == 0 {
		// elements in the values array are stored at their
		// relative position inside the map to maintain order
		//
		// input.key: [["foo","bar"],[123,456]]
		// 	cache[0][]interface{}{"foo",123}
		// 	cache[1][]interface{}{"bar",456}
		cache := make(map[int][]interface{})
		res := cap.Get(p.InputKey)
		for _, val := range res.Array() {
			for x, v := range val.Array() {
				cache[x] = append(cache[x], v.Value())
			}
		}

		var array []interface{}
		for i := 0; i < len(cache); i++ {
			array = append(array, cache[i])
		}

		// [["foo",123],["bar",456]]
		cap.Set(p.OutputKey, array)
		return cap, nil
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
	res := cap.Get(p.InputKey)
	for x, val := range res.Array() {
		for x1, v1 := range val.Array() {
			cache[x1], err = json.Set(cache[x1], p.Options.Keys[x], v1)
			if err != nil {
				return cap, fmt.Errorf("applicator settings %+v: %w", p, err)
			}
		}
	}

	// inserts pre-formatted JSON into an array based
	// on the length of the map
	var tmp []byte
	for i := 0; i < len(cache); i++ {
		tmp, err = json.Set(tmp, fmt.Sprintf("%d", i), cache[i])
		if err != nil {
			return cap, fmt.Errorf("applicator settings %+v: %w", p, err)
		}
	}

	// JSON arrays must be set using SetRaw to preserve structure
	// [{"name":"foo","size":123},{"name":"bar","size":456}]
	cap.SetRaw(p.OutputKey, tmp)
	return cap, nil
}
