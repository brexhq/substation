package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/json"
)

// group processes data by grouping object arrays into an array of tuples or array of objects.
//
// This processor supports the object handling pattern.
type procGroup struct {
	process
	Options procGroupOptions `json:"options"`
}

type procGroupOptions struct {
	// Keys determines where processed values are set in newly created objects.
	//
	// This is optional and defaults to creating an array of tuples instead
	// of an array of objects.
	Keys []string `json:"keys"`
}

// Create a new group processor.
func newProcGroup(ctx context.Context, cfg config.Config) (p procGroup, err error) {
	if err = config.Decode(cfg.Settings, &p); err != nil {
		return procGroup{}, err
	}

	p.operator, err = condition.NewOperator(ctx, p.Condition)
	if err != nil {
		return procGroup{}, err
	}

	// only supports JSON arrays, fail if there are no keys
	if p.Key == "" && p.SetKey == "" {
		return procGroup{}, fmt.Errorf("process: group: options %+v: %v", p.Options, errors.ErrMissingRequiredOption)
	}

	return p, nil
}

// String returns the processor settings as an object.
func (p procGroup) String() string {
	return toString(p)
}

// Closes resources opened by the processor.
func (p procGroup) Close(context.Context) error {
	return nil
}

// Stream processes a pipeline of capsules with the processor.
func (p procGroup) Stream(ctx context.Context, in, out *config.Channel) error {
	return streamApply(ctx, in, out, p)
}

// Batch processes one or more capsules with the processor.
func (p procGroup) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	return batchApply(ctx, capsules, p)
}

// Apply processes a capsule with the processor.
func (p procGroup) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	if ok, err := p.operator.Operate(ctx, capsule); err != nil {
		return capsule, fmt.Errorf("process: group: %v", err)
	} else if !ok {
		return capsule, nil
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
			return capsule, fmt.Errorf("process: group: %v", err)
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
				return capsule, fmt.Errorf("process: group: %v", err)
			}
		}
	}

	// inserts pre-formatted JSON into an array based
	// on the length of the map
	var value []byte
	for i := 0; i < len(cache); i++ {
		value, err = json.Set(value, fmt.Sprintf("%d", i), cache[i])
		if err != nil {
			return capsule, fmt.Errorf("process: group: %v", err)
		}
	}

	// JSON arrays must be set using SetRaw to preserve structure
	// [{"name":"foo","size":123},{"name":"bar","size":456}]
	if err := capsule.SetRaw(p.SetKey, value); err != nil {
		return capsule, fmt.Errorf("process: group: %v", err)
	}

	return capsule, nil
}
