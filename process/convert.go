package process

import (
	"context"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/internal/json"
)

/*
ConvertOptions contain custom options settings for this processor.

Type: the type that the value should be converted to; one of: bool (boolean), int (integer), float, uint (unsigned integer), string
*/
type ConvertOptions struct {
	Type string `mapstructure:"type"`
}

// Convert implements the Byter and Channeler interfaces and converts values between types. More information is available in the README.
type Convert struct {
	Condition condition.OperatorConfig `mapstructure:"condition"`
	Input     Input                    `mapstructure:"input"`
	Output    Output                   `mapstructure:"output"`
	Options   ConvertOptions           `mapstructure:"options"`
}

// Channel processes a data channel of bytes with this processor. Conditions can be optionally applied on the channel data to enable processing.
func (p Convert) Channel(ctx context.Context, ch <-chan []byte) (<-chan []byte, error) {
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
func (p Convert) Byte(ctx context.Context, data []byte) ([]byte, error) {
	value := json.Get(data, p.Input.Key)

	if !value.IsArray() {
		o := p.convert(value)
		return json.Set(data, p.Output.Key, o)
	}

	var array []interface{}
	for _, v := range value.Array() {
		o := p.convert(v)
		array = append(array, o)
	}

	return json.Set(data, p.Output.Key, array)
}

func (p Convert) convert(v json.Result) interface{} {
	switch t := p.Options.Type; t {
	case "bool":
		return v.Bool()
	case "int":
		return v.Int()
	case "float":
		return v.Float()
	case "uint":
		return v.Uint()
	case "string":
		return v.String()
	default:
		return nil
	}
}
