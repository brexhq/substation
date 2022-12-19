package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/json"
)

type group struct {
	process
	Options groupOptions `json:"options"`
}

type groupOptions struct {
	Keys []string `json:"keys"`
}

// Close closes resources opened by the group processor.
func (p group) Close(context.Context) error {
	return nil
}

func (p group) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	capsules, err := conditionalApply(ctx, capsules, p.Condition, p)
	if err != nil {
		return nil, fmt.Errorf("process group: %v", err)
	}

	return capsules, nil
}

// Apply processes encapsulated data with the group processor.
func (p group) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	// only supports JSON arrays, error early if there are no keys
	if p.Key == "" && p.SetKey == "" {
		return capsule, fmt.Errorf("process group: options %+v: %v", p.Options, errMissingRequiredOptions)
	}

	if len(p.Options.Keys) == 0 {
		// elements in the values array are stored at their
		// relative position inside the map to maintain order
		//
		// input.key: [["foo","bar"],[123,456]]
		// 	cache[0][]interface{}{"foo",123}
		// 	cache[1][]interface{}{"bar",456}
		cache := make(map[int][]interface{})
		result := capsule.Get(p.Key)
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
		if err := capsule.Set(p.SetKey, value); err != nil {
			return capsule, fmt.Errorf("process group: %v", err)
		}

		return capsule, nil
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
	result := capsule.Get(p.Key)
	for i, res := range result.Array() {
		for j, r := range res.Array() {
			cache[j], err = json.Set(cache[j], p.Options.Keys[i], r)
			if err != nil {
				return capsule, fmt.Errorf("process group: %v", err)
			}
		}
	}

	// inserts pre-formatted JSON into an array based
	// on the length of the map
	var value []byte
	for i := 0; i < len(cache); i++ {
		value, err = json.Set(value, fmt.Sprintf("%d", i), cache[i])
		if err != nil {
			return capsule, fmt.Errorf("process group: %v", err)
		}
	}

	// JSON arrays must be set using SetRaw to preserve structure
	// [{"name":"foo","size":123},{"name":"bar","size":456}]
	if err := capsule.SetRaw(p.SetKey, value); err != nil {
		return capsule, fmt.Errorf("process group: %v", err)
	}

	return capsule, nil
}
