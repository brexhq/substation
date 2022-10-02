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
		return nil, fmt.Errorf("process group: %v", err)
	}

	caps, err = conditionallyApplyBatch(ctx, caps, op, p)
	if err != nil {
		return nil, fmt.Errorf("process group: %v", err)
	}

	return caps, nil
}

// Apply processes encapsulated data with the Group processor.
func (p Group) Apply(ctx context.Context, cap config.Capsule) (config.Capsule, error) {
	// only supports JSON arrays, error early if there are no keys
	if p.InputKey == "" && p.OutputKey == "" {
		return cap, fmt.Errorf("process group: options %+v: %v", p.Options, errProcessorMissingRequiredOptions)
	}

	if len(p.Options.Keys) == 0 {
		// elements in the values array are stored at their
		// relative position inside the map to maintain order
		//
		// input.key: [["foo","bar"],[123,456]]
		// 	cache[0][]interface{}{"foo",123}
		// 	cache[1][]interface{}{"bar",456}
		cache := make(map[int][]interface{})
		result := cap.Get(p.InputKey)
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
		if err := cap.Set(p.OutputKey, value); err != nil {
			return cap, fmt.Errorf("process group: %v", err)
		}

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
	result := cap.Get(p.InputKey)
	for i, res := range result.Array() {
		for j, r := range res.Array() {
			cache[j], err = json.Set(cache[j], p.Options.Keys[i], r)
			if err != nil {
				return cap, fmt.Errorf("process group: %v", err)
			}
		}
	}

	// inserts pre-formatted JSON into an array based
	// on the length of the map
	var value []byte
	for i := 0; i < len(cache); i++ {
		value, err = json.Set(value, fmt.Sprintf("%d", i), cache[i])
		if err != nil {
			return cap, fmt.Errorf("process group: %v", err)
		}
	}

	// JSON arrays must be set using SetRaw to preserve structure
	// [{"name":"foo","size":123},{"name":"bar","size":456}]
	if err := cap.SetRaw(p.OutputKey, value); err != nil {
		return cap, fmt.Errorf("process group: %v", err)
	}

	return cap, nil
}
