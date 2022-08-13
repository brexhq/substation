package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/json"
)

/*
Expand processes encapsulated data by creating individual events from objects in JSON arrays. The processor supports these patterns:
	JSON:
		{"expand":[{"foo":"bar"}],"baz":"qux"} >>> {"foo":"bar","baz":"qux"}

The processor uses this Jsonnet configuration:
	{
		type: 'expand',
		settings: {
			input_key: 'expand',
		},
	}
*/
type Expand struct {
	Condition condition.OperatorConfig `json:"condition"`
	InputKey  string                   `json:"input_key"`
}

// ApplyBatch processes a slice of encapsulated data with the Expand processor. Conditions are optionally applied to the data to enable processing.
func (p Expand) ApplyBatch(ctx context.Context, caps []config.Capsule) ([]config.Capsule, error) {
	// only supports JSON, error early if there is no input key
	if p.InputKey == "" {
		return nil, fmt.Errorf("applybatch settings %+v: %w", p, ProcessorInvalidSettings)
	}

	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, fmt.Errorf("applybatch settings %+v: %w", p, err)
	}

	newCaps := NewBatch(&caps)
	for _, cap := range caps {
		ok, err := op.Operate(cap)
		if err != nil {
			return nil, fmt.Errorf("applybatch settings %+v: %w", p, err)
		}

		if !ok {
			newCaps = append(newCaps, cap)
			continue
		}

		// data is processed by retrieving and iterating the
		// array (InputKey) containing JSON objects and setting
		// any additional keys from the root object into each
		// expanded object
		//
		// root:
		// 	{"expand":[{"foo":"bar"},{"baz":"qux"}],"quux":"corge"}
		// expanded:
		// 	{"foo":"bar","quux":"corge"}
		// 	{"baz":"qux","quux":"corge"}
		root := cap.Get("@this")
		// root := json.Get(data, "@this")
		res := cap.Get(p.InputKey)
		// v := json.Get(data, p.InputKey)
		for _, result := range res.Array() {
			var err error

			expand := []byte(result.String())
			for key, val := range root.Map() {
				if key == p.InputKey {
					continue
				}

				expand, err = json.Set(expand, key, val)
				if err != nil {
					return nil, fmt.Errorf("applybatch settings %+v: %w", p, err)
				}
			}

			newCap := config.NewCapsule()
			newCap.SetData(expand)

			newCaps = append(newCaps, newCap)
		}
	}

	return newCaps, nil
}
