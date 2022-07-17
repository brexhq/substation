package process

import (
	"context"
	"crypto/md5"
	"crypto/sha256"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/json"
)

// HashUnsupportedAlgorithm is returned when the Hash processor is configured with an unsupported algorithm.
const HashUnsupportedAlgorithm = errors.Error("HashUnsupportedAlgorithm")

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
	json:
		{"hash":"foo"} >>> {"hash":"acbd18db4cc2f85cedef654fccc4a4d8"}
	data:
		foo >>> acbd18db4cc2f85cedef654fccc4a4d8

The processor uses this Jsonnet configuration:
	{
		type: 'hash',
		settings: {
			input_key: 'hash',
			output_key: 'hash',
			options: {
				algorithm: 'md5',
			}
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

// Byte processes bytes with the Hash processor.
func (p Hash) Byte(ctx context.Context, data []byte) ([]byte, error) {
	// error early if required options are missing
	if p.Options.Algorithm == "" {
		return nil, fmt.Errorf("byter settings %+v: %v", p, ProcessorInvalidSettings)
	}

	// JSON processing
	if p.InputKey != "" && p.OutputKey != "" {
		value := json.Get(data, p.InputKey)
		b := []byte(value.String())
		h, err := p.hash(b)
		if err != nil {
			return nil, fmt.Errorf("byter settings %+v: %v", p, err)
		}

		return json.Set(data, p.OutputKey, h)
	}

	// data processing
	if p.InputKey == "" && p.OutputKey == "" {
		h, err := p.hash(data)
		if err != nil {
			return nil, fmt.Errorf("byter settings %+v: %v", p, err)
		}
		return []byte(h), nil
	}

	return nil, fmt.Errorf("byter settings %+v: %v", p, ProcessorInvalidSettings)
}

func (p Hash) hash(b []byte) (string, error) {
	switch s := p.Options.Algorithm; s {
	case "md5":
		sum := md5.Sum(b)
		return fmt.Sprintf("%x", sum), nil
	case "sha256":
		sum := sha256.Sum256(b)
		return fmt.Sprintf("%x", sum), nil
	default:
		return "", fmt.Errorf("hash type %s: %v", s, HashUnsupportedAlgorithm)
	}
}
