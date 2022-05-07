package process

import (
	"context"

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
*/
type MathOptions struct {
	Operation string `json:"operation"`
}

/*
Math processes data by applying mathematic operations. The processor supports these patterns:
	json:
		{"foo":1,"bar":3} >>> {"foo":1,"bar":3,"math":4}
	json array:
		{"foo":[1,2],"bar":[3,4]} >>> {"foo":[1,2],"bar":[3,4],"math":[4,6]}

The processor uses this Jsonnet configuration:
	{
		type: 'math',
		settings: {
			input: {
				keys: ['foo','bar'],
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
	Input     Inputs                   `json:"input"`
	Output    Output                   `json:"output"`
	Options   MathOptions              `json:"options"`
}

// Slice processes a slice of bytes with the Math processor. Conditions are optionally applied on the bytes to enable processing.
func (p Math) Slice(ctx context.Context, s [][]byte) ([][]byte, error) {
	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, err
	}

	slice := NewSlice(&s)
	for _, data := range s {
		ok, err := op.Operate(data)
		if err != nil {
			return nil, err
		}

		if !ok {
			slice = append(slice, data)
			continue
		}

		processed, err := p.Byte(ctx, data)
		if err != nil {
			return nil, err
		}
		slice = append(slice, processed)
	}

	return slice, nil
}

// Byte processes bytes with the Math processor.
func (p Math) Byte(ctx context.Context, data []byte) ([]byte, error) {
	// only supports json and json arrays, so error early if there are no keys
	if len(p.Input.Keys) == 0 && p.Output.Key == "" {
		return nil, MathInvalidSettings
	}

	// simultaneously processes json and json arrays
	cache := make(map[int]int64)
	for i, key := range p.Input.Keys {
		value := json.Get(data, key)

		for x, v := range value.Array() {
			if i == 0 {
				cache[x] = v.Int()
				continue
			}

			switch p.Options.Operation {
			case "add":
				cache[x] = cache[x] + v.Int()
			case "subtract":
				cache[x] = cache[x] - v.Int()
			}
		}
	}

	if len(cache) == 1 {
		return json.Set(data, p.Output.Key, cache[0])
	}

	var array []int64
	for _, v := range cache {
		array = append(array, v)
	}

	return json.Set(data, p.Output.Key, array)
}
