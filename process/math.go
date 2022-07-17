package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/internal/json"
)

/*
MathOptions contains custom options for the Math processor:
	Operation:
		the operator applied to the data
		must be one of:
			add
			subtract
			divide
*/
type MathOptions struct {
	Operation string `json:"operation"`
}

/*
Math processes data by applying mathematic operations. The processor supports these patterns:
	json:
		{"math":[1,3]} >>> {"math":4}

The processor uses this Jsonnet configuration:
	{
		type: 'math',
		settings: {
			input_key: 'math',
			output_key: 'math',
			options: {
				operation: 'add',
			}
		},
	}
*/
type Math struct {
	Options   MathOptions              `json:"options"`
	Condition condition.OperatorConfig `json:"condition"`
	InputKey  string                   `json:"input_key"`
	OutputKey string                   `json:"output_key"`
}

// Slice processes a slice of bytes with the Math processor. Conditions are optionally applied on the bytes to enable processing.
func (p Math) Slice(ctx context.Context, s [][]byte) ([][]byte, error) {
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

// Byte processes bytes with the Math processor.
func (p Math) Byte(ctx context.Context, data []byte) ([]byte, error) {
	// error early if required options are missing
	if p.Options.Operation == "" {
		return nil, fmt.Errorf("byter settings %+v: %v", p, ProcessorInvalidSettings)
	}

	// only supports json, error early if there are no keys
	if p.InputKey == "" && p.OutputKey == "" {
		return nil, fmt.Errorf("byter settings %+v: %v", p, ProcessorInvalidSettings)
	}

	// elements in the values array are stored at their
	// 	relative position inside the map to maintain order
	//
	// input.key: [[1,2],[6,10]]
	// options.operation: add
	// 	cache[0:7]
	// 	cache[1:12]
	cache := make(map[int]int64)
	value := json.Get(data, p.InputKey)
	for x, v := range value.Array() {
		var idx int

		for x1, v1 := range v.Array() {
			if v.IsArray() {
				idx = x1
			}

			if x == 0 {
				cache[idx] = v1.Int()
				continue
			}

			switch p.Options.Operation {
			case "add":
				cache[idx] = cache[idx] + v1.Int()
			case "subtract":
				cache[idx] = cache[idx] - v1.Int()
			case "divide":
				cache[idx] = cache[idx] / v1.Int()
			}
		}
	}

	if len(cache) == 1 {
		return json.Set(data, p.OutputKey, cache[0])
	}

	var array []int64
	for i := 0; i < len(cache); i++ {
		array = append(array, cache[i])
	}

	return json.Set(data, p.OutputKey, array)
}
