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
HashOptions contain custom options settings for this processor.

Algorithm: the algorithm to apply.
*/
type HashOptions struct {
	Algorithm string `mapstructure:"algorithm"`
}

// Hash implements the Byter and Channeler interfaces and calculates the hash of data. More information is available in the README.
type Hash struct {
	Condition condition.OperatorConfig `mapstructure:"condition"`
	Input     Input                    `mapstructure:"input"`
	Output    Output                   `mapstructure:"output"`
	Options   HashOptions              `mapstructure:"options"`
}

// Channel processes a data channel of bytes with this processor. Conditions can be optionally applied on the channel data to enable processing.
func (p Hash) Channel(ctx context.Context, ch <-chan []byte) (<-chan []byte, error) {
	var array [][]byte

	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, err
	}

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

// Byte processes a byte slice with this processor
func (p Hash) Byte(ctx context.Context, data []byte) ([]byte, error) {
	value := json.Get(data, p.Input.Key)

	if !value.IsArray() {
		b := []byte(value.String())
		o := p.hash(b)
		return json.Set(data, p.Output.Key, o)
	}

	var array []string
	for _, v := range value.Array() {
		b := []byte(v.String())
		o := p.hash(b)
		array = append(array, o)
	}

	return json.Set(data, p.Output.Key, array)
}

func (p Hash) hash(b []byte) string {
	switch a := p.Options.Algorithm; a {
	case "md5":
		sum := md5.Sum(b)
		return fmt.Sprintf("%x", sum)
	case "sha256":
		sum := sha256.Sum256(b)
		return fmt.Sprintf("%x", sum)
	default:
		return ""
	}
}
