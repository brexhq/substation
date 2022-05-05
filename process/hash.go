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

// HashInvalidSettings is returned when the Hash processor is configured with invalid Input and Output settings.
const HashInvalidSettings = errors.Error("HashInvalidSettings")

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
	Algorithm string `mapstructure:"algorithm"`
}

/*
Hash processes data by calculating hashes. The processor supports these patterns:
	json:
		{"hash":"foo"} >>> {"hash":"acbd18db4cc2f85cedef654fccc4a4d8"}
	json array:
		{"hash":["foo","bar"]} >>> {"hash":["acbd18db4cc2f85cedef654fccc4a4d8","37b51d194a7513e45b56f6524f2d51f2"]}
	data:
		foo >>> acbd18db4cc2f85cedef654fccc4a4d8

The processor uses this Jsonnet configuration:
	{
		type: 'hash',
		settings: {
			input: {
				key: 'hash',
			},
			output: {
				key: 'hash',
			}
			options: {
				algorithm: 'md5',
			}
		},
	}
*/
type Hash struct {
	Condition condition.OperatorConfig `mapstructure:"condition"`
	Input     Input                    `mapstructure:"input"`
	Output    Output                   `mapstructure:"output"`
	Options   HashOptions              `mapstructure:"options"`
}

// Channel processes a data channel of byte slices with the Hash processor. Conditions are optionally applied on the channel data to enable processing.
func (p Hash) Channel(ctx context.Context, ch <-chan []byte) (<-chan []byte, error) {
	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, err
	}

	var array [][]byte
	for data := range ch {
		ok, err := op.Operate(data)
		if err != nil {
			return nil, err
		}

		if !ok {
			array = append(array, data)
			continue
		}

		processed, err := p.Byte(ctx, data)
		if err != nil {
			return nil, err
		}
		array = append(array, processed)
	}

	output := make(chan []byte, len(array))
	for _, x := range array {
		output <- x
	}
	close(output)
	return output, nil

}

// Byte processes a byte slice with the Hash processor.
func (p Hash) Byte(ctx context.Context, data []byte) ([]byte, error) {
	// json processing
	if p.Input.Key != "" && p.Output.Key != "" {
		value := json.Get(data, p.Input.Key)
		if !value.IsArray() {
			b := []byte(value.String())
			h, err := p.hash(b)
			if err != nil {
				return nil, err
			}

			return json.Set(data, p.Output.Key, h)
		}

		// json array processing
		var array []string
		for _, v := range value.Array() {
			b := []byte(v.String())
			h, err := p.hash(b)
			if err != nil {
				return nil, err
			}

			array = append(array, h)
		}

		return json.Set(data, p.Output.Key, array)
	}

	// data processing
	if p.Input.Key == "" && p.Output.Key == "" {
		h, err := p.hash(data)
		if err != nil {
			return nil, err
		}
		return []byte(h), nil
	}

	return nil, HashInvalidSettings
}

func (p Hash) hash(b []byte) (string, error) {
	switch a := p.Options.Algorithm; a {
	case "md5":
		sum := md5.Sum(b)
		return fmt.Sprintf("%x", sum), nil
	case "sha256":
		sum := sha256.Sum256(b)
		return fmt.Sprintf("%x", sum), nil
	default:
		return "", HashUnsupportedAlgorithm
	}
}
