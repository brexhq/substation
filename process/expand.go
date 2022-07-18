package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/internal/json"
)

/*
Expand processes data by creating individual events from objects in JSON arrays. The processor supports these patterns:
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

// Slice processes a slice of bytes with the Expand processor. Conditions are optionally applied on the bytes to enable processing.
func (p Expand) Slice(ctx context.Context, s [][]byte) ([][]byte, error) {
	// only supports JSON, error early if there is no input key
	if p.InputKey == "" {
		return nil, fmt.Errorf("slicer settings %+v: %w", p, ProcessorInvalidSettings)
	}

	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, fmt.Errorf("slicer settings %+v: %w", p, err)
	}

	slice := NewSlice(&s)
	for _, data := range s {
		ok, err := op.Operate(data)
		if err != nil {
			return nil, fmt.Errorf("slicer settings %+v: %w", p, err)
		}

		if !ok {
			slice = append(slice, data)
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
		root := json.Get(data, "@this")
		v := json.Get(data, p.InputKey)
		for _, value := range v.Array() {
			var err error

			expand := []byte(value.String())
			for k, v := range root.Map() {
				if k == p.InputKey {
					continue
				}

				expand, err = json.Set(expand, k, v)
				if err != nil {
					return nil, fmt.Errorf("slicer settings %+v: %w", p, err)
				}
			}

			slice = append(slice, expand)
		}
	}

	return slice, nil
}
