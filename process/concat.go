package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/json"
)

// ConcatInvalidSettings is returned when the Concat processor is configured with invalid Input and Output settings.
const ConcatInvalidSettings = errors.Error("ConcatInvalidSettings")

/*
ConcatOptions contains custom options for the Concat processor:
	Separator:
		the string that separates the concatenated values
*/
type ConcatOptions struct {
	Separator string `json:"separator"`
}

/*
Concat processes data by concatenating multiple values together with a separator. The processor supports these patterns:
	json:
		{"concat":["foo","bar"]} >>> {"concat":"foo.bar"}
	json array:
		{"concat":[["foo","baz"],["bar","qux"]]} >>> {"concat":["foo.bar","baz.qux"]}

The processor uses this Jsonnet configuration:
	{
		type: 'concat',
		settings: {
			input: {
				key: 'concat',
			},
			output: {
				key: 'concat',
			},
			options: {
				separator: '.',
			}
		},
	}
*/
type Concat struct {
	Condition condition.OperatorConfig `json:"condition"`
	Input     string                   `json:"input"`
	Output    string                   `json:"output"`
	Options   ConcatOptions            `json:"options"`
}

// Slice processes a slice of bytes with the Concat processor. Conditions are optionally applied on the bytes to enable processing.
func (p Concat) Slice(ctx context.Context, s [][]byte) ([][]byte, error) {
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

// Byte processes bytes with the Concat processor.
func (p Concat) Byte(ctx context.Context, data []byte) ([]byte, error) {
	// only supports json and json arrays, error early if there are no keys
	if p.Input == "" && p.Output == "" {
		return nil, fmt.Errorf("byter settings %v: %v", p, ConcatInvalidSettings)
	}

	cache := make(map[int]string)
	value := json.Get(data, p.Input)
	for x, v := range value.Array() {
		var idx int

		for x1, v1 := range v.Array() {
			if v.IsArray() {
				idx = x1
			}

			cache[idx] += v1.String()
			if x != len(value.Array())-1 {
				cache[idx] += p.Options.Separator
			}
		}
	}

	// json processing
	if len(cache) == 1 {
		return json.Set(data, p.Output, cache[0])
	}

	// json array processing
	var array []string
	for i := 0; i < len(cache); i++ {
		array = append(array, cache[i])
	}

	return json.Set(data, p.Output, array)
}
