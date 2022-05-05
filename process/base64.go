package process

import (
	"context"
	"encoding/base64"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/json"
)

// Base64InvalidSettings is returned when the Base64 processor is configured with invalid Input and Output settings.
const Base64InvalidSettings = errors.Error("Base64InvalidSettings")

// Base64InvalidDirection is returned when the Base64 processor is configured with an invalid direction setting.
const Base64InvalidDirection = errors.Error("Base64InvalidDirection")

// Base64InvalidAlphabet is returned when the Base64 processor is configured with an invalid alphabet setting.
const Base64InvalidAlphabet = errors.Error("Base64InvalidAlphabet")

/*
Base64Options contains custom options for the Base64 processor:
	Direction:
		the direction of the encoding
		must be one of:
			to: encode to base64
			from: decode from base64
	Alphabet:
		the base64 alphabet to use, either std (https://www.rfc-editor.org/rfc/rfc4648.html#section-4) or url (https://www.rfc-editor.org/rfc/rfc4648.html#section-5)
		defaults to std
*/
type Base64Options struct {
	Direction string `mapstructure:"direction"`
	Alphabet  string `mapstructure:"alphabet"`
}

/*
Base64 processes data by converting it to and from base64. The processor supports these patterns:
	json:
  	{"base64":"Zm9v"} >>> {"base64":"foo"}
	json array:
		{"base64":["Zm9v","YmFy"]} >>> {"base64":["foo","bar"]}
	data:
		Zm9v >>> foo

The processor uses this Jsonnet configuration:
	{
		type: 'base64',
		settings: {
			input: {
				key: 'base64',
			},
			output: {
				key: 'base64',
			},
			options: {
				direction: 'from',
			}
		},
	}
*/
type Base64 struct {
	Condition condition.OperatorConfig `mapstructure:"condition"`
	Input     Input                    `mapstructure:"input"`
	Output    Output                   `mapstructure:"output"`
	Options   Base64Options            `mapstructure:"options"`
}

// Channel processes a data channel of byte slices with the Base64 processor. Conditions are optionally applied on the channel data to enable processing.
func (p Base64) Channel(ctx context.Context, ch <-chan []byte) (<-chan []byte, error) {
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

// Byte processes a byte slice with the Base64 processor.
func (p Base64) Byte(ctx context.Context, data []byte) ([]byte, error) {
	if p.Options.Alphabet == "" {
		p.Options.Alphabet = "std"
	}

	// json processing
	if p.Input.Key != "" && p.Output.Key != "" {
		value := json.Get(data, p.Input.Key)

		if !value.IsArray() {
			tmp := []byte(value.String())
			if p.Options.Direction == "from" {
				result, err := p.from(tmp, p.Options.Alphabet)
				if err != nil {
					return nil, err
				}
				return json.Set(data, p.Output.Key, result)
			} else if p.Options.Direction == "to" {
				result, err := p.to(tmp, p.Options.Alphabet)
				if err != nil {
					return nil, err
				}
				return json.Set(data, p.Output.Key, result)
			} else {
				return nil, Base64InvalidDirection
			}
		}

		// json array processing
		var array []string
		for _, v := range value.Array() {
			tmp := []byte(v.String())
			if p.Options.Direction == "from" {
				result, err := p.from(tmp, p.Options.Alphabet)
				if err != nil {
					return nil, err
				}
				array = append(array, string(result))
			} else if p.Options.Direction == "to" {
				result, err := p.to(tmp, p.Options.Alphabet)
				if err != nil {
					return nil, err
				}
				array = append(array, string(result))
			} else {
				return nil, Base64InvalidDirection
			}
		}

		return json.Set(data, p.Output.Key, array)
	}

	// data processing
	if p.Input.Key == "" && p.Output.Key == "" {
		if p.Options.Direction == "from" {
			result, err := p.from(data, p.Options.Alphabet)
			if err != nil {
				return nil, err
			}
			return result, nil
		} else if p.Options.Direction == "to" {
			result, err := p.to(data, p.Options.Alphabet)
			if err != nil {
				return nil, err
			}
			return result, nil
		} else {
			return nil, Base64InvalidDirection
		}
	}

	return nil, Base64InvalidSettings
}

func (p Base64) from(data []byte, alphabet string) ([]byte, error) {
	len := len(string(data))

	switch s := alphabet; s {
	case "std":
		res := make([]byte, base64.StdEncoding.DecodedLen(len))
		n, err := base64.StdEncoding.Decode(res, data)
		if err != nil {
			return nil, err
		}

		return res[:n], nil
	case "url":
		res := make([]byte, base64.URLEncoding.DecodedLen(len))
		n, err := base64.URLEncoding.Decode(res, data)
		if err != nil {
			return nil, err
		}

		return res[:n], nil
	default:
		return nil, Base64InvalidAlphabet
	}
}

func (p Base64) to(data []byte, alphabet string) ([]byte, error) {
	len := len(data)

	switch s := alphabet; s {
	case "std":
		res := make([]byte, base64.StdEncoding.EncodedLen(len))
		base64.StdEncoding.Encode(res, data)
		return res, nil
	case "url":
		res := make([]byte, base64.URLEncoding.EncodedLen(len))
		base64.URLEncoding.Encode(res, data)
		return res, nil
	default:
		return nil, Base64InvalidAlphabet
	}
}
