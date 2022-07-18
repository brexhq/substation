package process

import (
	"context"
	"crypto/md5"
	"crypto/sha256"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/internal/json"
)

/*
HashOptions contains custom options for the Hash processor:
	Algorithm:
		the hashing algorithm to apply
		must be one of:
			md5
			sha256
*/
type HashOptions struct {
	Algorithm string `json:"algorithm"`
}

/*
Hash processes data by calculating hashes. The processor supports these patterns:
	JSON:
		{"hash":"foo"} >>> {"hash":"acbd18db4cc2f85cedef654fccc4a4d8"}
	data:
		foo >>> acbd18db4cc2f85cedef654fccc4a4d8

The processor uses this Jsonnet configuration:
	{
		type: 'hash',
		settings: {
			options: {
				algorithm: 'md5',
			},
			input_key: 'hash',
			output_key: 'hash',
		},
	}
*/
type Hash struct {
	Options   HashOptions              `json:"options"`
	Condition condition.OperatorConfig `json:"condition"`
	InputKey  string                   `json:"input_key"`
	OutputKey string                   `json:"output_key"`
}

// Slice processes a slice of bytes with the Hash processor. Conditions are optionally applied on the bytes to enable processing.
func (p Hash) Slice(ctx context.Context, s [][]byte) ([][]byte, error) {
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

// Byte processes bytes with the Hash processor.
func (p Hash) Byte(ctx context.Context, data []byte) ([]byte, error) {
	// error early if required options are missing
	if p.Options.Algorithm == "" {
		return nil, fmt.Errorf("byter settings %+v: %w", p, ProcessorInvalidSettings)
	}

	// JSON processing
	if p.InputKey != "" && p.OutputKey != "" {
		value := json.Get(data, p.InputKey).String()
		switch p.Options.Algorithm {
		case "md5":
			sum := md5.Sum([]byte(value))
			return json.Set(data, p.OutputKey, fmt.Sprintf("%x", sum))
		case "sha256":
			sum := sha256.Sum256([]byte(value))
			return json.Set(data, p.OutputKey, fmt.Sprintf("%x", sum))
		}
	}

	// data processing
	if p.InputKey == "" && p.OutputKey == "" {
		switch p.Options.Algorithm {
		case "md5":
			sum := md5.Sum(data)
			x := fmt.Sprintf("%x", sum)
			return []byte(x), nil
		case "sha256":
			sum := sha256.Sum256(data)
			x := fmt.Sprintf("%x", sum)
			return []byte(x), nil
		}
	}

	return nil, fmt.Errorf("byter settings %+v: %w", p, ProcessorInvalidSettings)
}
