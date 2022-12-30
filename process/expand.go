package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/json"
)

// expand processes data by creating new objects from objects in arrays.
//
// This processor supports the data and object handling patterns.
type _expand struct {
	process
}

// String returns the processor settings as an object.
func (p _expand) String() string {
	return toString(p)
}

// Close closes resources opened by the processor.
func (p _expand) Close(context.Context) error {
	return nil
}

// Batch processes one or more capsules with the processor. Conditions are
// optionally applied to the data to enable processing.
func (p _expand) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	op, err := condition.MakeOperator(p.Condition)
	if err != nil {
		return nil, fmt.Errorf("process: expand: %v", err)
	}

	newCapsules := newBatch(&capsules)
	for _, capsule := range capsules {
		ok, err := op.Operate(ctx, capsule)
		if err != nil {
			return nil, fmt.Errorf("process: expand: %v", err)
		}

		if !ok {
			newCapsules = append(newCapsules, capsule)
			continue
		}

		// data is processed by retrieving and iterating the
		// array containing JSON objects and setting
		// any additional keys from the root object into each
		// expanded object. if there is no Key, then the
		// input is processed as an array.
		//
		// root:
		// 	{"_expand":[{"foo":"bar"},{"baz":"qux"}],"quux":"corge"}
		// expanded:
		// 	{"foo":"bar","quux":"corge"}
		// 	{"baz":"qux","quux":"corge"}
		root := capsule.Get("@this")
		result := root

		// JSON processing
		// the Get / Delete routine is a hack to speed up processing
		// very large objects, like those output by AWS CloudTrail.
		if p.Key != "" {
			rootBytes, err := json.Delete([]byte(root.String()), p.Key)
			if err != nil {
				return nil, fmt.Errorf("process: expand: %v", err)
			}

			root = json.Get(rootBytes, "@this")
			result = capsule.Get(p.Key)
		}

		// retains metadata from the original capsule
		newCapsule := capsule
		for _, res := range result.Array() {
			var err error

			_expand := []byte(res.String())
			for key, val := range root.Map() {
				if key == p.Key {
					continue
				}

				_expand, err = json.Set(_expand, key, val)
				if err != nil {
					return nil, fmt.Errorf("process: expand: %v", err)
				}
			}

			newCapsule.SetData(_expand)
			newCapsules = append(newCapsules, newCapsule)
		}
	}

	return newCapsules, nil
}
