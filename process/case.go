package process

import (
	"context"
	"strings"

	"github.com/iancoleman/strcase"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/internal/json"
)

/*
CaseOptions contain custom options settings for this processor.

Case: the case to convert the string to; one of: upper, lower, or snake
*/
type CaseOptions struct {
	Case string `mapstructure:"case"`
}

// Case implements the Byter and Channeler interfaces and converts the case of a string. More information is available in the README.
type Case struct {
	Condition condition.OperatorConfig `mapstructure:"condition"`
	Input     Input                    `mapstructure:"input"`
	Output    Output                   `mapstructure:"output"`
	Options   CaseOptions              `mapstructure:"options"`
}

// Channel processes a data channel of bytes with this processor. Conditions can be optionally applied on the channel data to enable processing.
func (p Case) Channel(ctx context.Context, ch <-chan []byte) (<-chan []byte, error) {
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
func (p Case) Byte(ctx context.Context, data []byte) ([]byte, error) {
	value := json.Get(data, p.Input.Key)

	if !value.IsArray() {
		s := value.String()
		c := p._case(s)
		return json.Set(data, p.Output.Key, c)
	}

	var array []string
	for _, v := range value.Array() {
		s := v.String()
		c := p._case(s)
		array = append(array, c)
	}

	return json.Set(data, p.Output.Key, array)
}

func (p Case) _case(s string) string {
	switch t := p.Options.Case; t {
	case "upper":
		return strings.ToUpper(s)
	case "lower":
		return strings.ToLower(s)
	case "snake":
		return strcase.ToSnake(s)
	default:
		return ""
	}
}
