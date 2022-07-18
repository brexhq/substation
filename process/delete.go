package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/json"
)

// DeleteInvalidSettings is returned when the Copy processor is configured with invalid Input and Output settings.
const DeleteInvalidSettings = errors.Error("DeleteInvalidSettings")

/*
Delete processes data by deleting JSON keys. The processor supports these patterns:
	JSON:
	  	{"foo":"bar","baz":"qux"} >>> {"foo":"bar"}

The processor uses this Jsonnet configuration:
	{
		type: 'delete',
		settings: {
			input_key: 'delete',
		}
	}
*/
type Delete struct {
	Condition condition.OperatorConfig `json:"condition"`
	InputKey  string                   `json:"input_key"`
}

// Slice processes a slice of bytes with the Delete processor. Conditions are optionally applied on the bytes to enable processing.
func (p Delete) Slice(ctx context.Context, s [][]byte) ([][]byte, error) {
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

		processed, err := p.Byte(ctx, data)
		if err != nil {
			return nil, fmt.Errorf("slicer: %v", err)
		}
		slice = append(slice, processed)
	}

	return slice, nil
}

// Byte processes bytes with the Delete processor.
func (p Delete) Byte(ctx context.Context, object []byte) ([]byte, error) {
	// only supports JSON, error early if there are no keys
	if p.InputKey == "" {
		return nil, fmt.Errorf("byter settings %+v: %w", p, ProcessorInvalidSettings)
	}

	return json.Delete(object, p.InputKey)
}
