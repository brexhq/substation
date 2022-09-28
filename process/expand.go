package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/json"
)

/*
Expand processes data by creating individual events from objects in arrays. The processor supports these patterns:
	JSON:
		{"expand":[{"foo":"bar"}],"baz":"qux"} >>> {"foo":"bar","baz":"qux"}
	data:
		[{"foo":"bar"}] >>> {"foo":"bar"}

When loaded with a factory, the processor uses this JSON configuration:
	{
		"type": "expand",
		"settings": {
			"input_key": "expand"
		}
	}
*/
type Expand struct {
	Condition condition.Config `json:"condition"`
	InputKey  string           `json:"input_key"`
}

// ApplyBatch processes a slice of encapsulated data with the Expand processor. Conditions are optionally applied to the data to enable processing.
func (p Expand) ApplyBatch(ctx context.Context, caps []config.Capsule) ([]config.Capsule, error) {
	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, fmt.Errorf("process expand applybatch: %v", err)
	}

	newCaps := newBatch(&caps)
	for _, cap := range caps {
		ok, err := op.Operate(ctx, cap)
		if err != nil {
			return nil, fmt.Errorf("process expand applybatch: %v", err)
		}

		if !ok {
			newCaps = append(newCaps, cap)
			continue
		}

		// data is processed by retrieving and iterating the
		// array containing JSON objects and setting
		// any additional keys from the root object into each
		// expanded object. if there is no InputKey, then the
		// input is processed as an array.
		//
		// root:
		// 	{"expand":[{"foo":"bar"},{"baz":"qux"}],"quux":"corge"}
		// expanded:
		// 	{"foo":"bar","quux":"corge"}
		// 	{"baz":"qux","quux":"corge"}
		root := cap.Get("@this")
		result := root

		// JSON processing
		// the Get / Delete routine is a hack to speed up processing
		// very large objects, like those output by AWS CloudTrail.
		if p.InputKey != "" {
			rootBytes, err := json.Delete([]byte(root.String()), p.InputKey)
			if err != nil {
				return nil, fmt.Errorf("process expand applybatch: %v", err)
			}

			root = json.Get(rootBytes, "@this")
			result = cap.Get(p.InputKey)
		}

		// retains metadata from the original capsule
		newCap := cap
		for _, res := range result.Array() {
			var err error

			expand := []byte(res.String())
			for key, val := range root.Map() {
				if key == p.InputKey {
					continue
				}

				expand, err = json.Set(expand, key, val)
				if err != nil {
					return nil, fmt.Errorf("process expand applybatch: %v", err)
				}
			}

			newCap.SetData(expand)
			newCaps = append(newCaps, newCap)
		}
	}

	return newCaps, nil
}
