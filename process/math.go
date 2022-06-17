package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/json"
)

// MathInvalidSettings is returned when the Math processor is configured with invalid Input and Output settings.
const MathInvalidSettings = errors.Error("MathInvalidSettings")

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
	json array:
		{"math":[[1,2],[3,4]]} >>> {"math":[4,6]}

The processor uses this Jsonnet configuration:
	{
		type: 'math',
		settings: {
			input: {
				key: 'math',
			},
			output: {
				key: 'math',
			}
			options: {
				operation: 'add',
			}
		},
	}
*/
type Math struct {
	Condition condition.OperatorConfig `json:"condition"`
	Input     string                   `json:"input"`
	Output    string                   `json:"output"`
	Options   MathOptions              `json:"options"`
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
	// only supports json and json arrays, error early if there are no keys
	if p.Input == "" && p.Output == "" {
		return nil, fmt.Errorf("byter settings %v: %v", p, MathInvalidSettings)
	}

	// elements in the values array are stored at their
	// 	relative position inside the map to maintain order
	//
	// input.key: [[1,2],[6,10]]
	// options.operation: add
	// 	cache[0:7]
	// 	cache[1:12]
	cache := make(map[int]int64)
	value := json.Get(data, p.Input)
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
		return json.Set(data, p.Output, cache[0])
	}

	var array []int64
	for i := 0; i < len(cache); i++ {
		array = append(array, cache[i])
	}

	return json.Set(data, p.Output, array)
}
