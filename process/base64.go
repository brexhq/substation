package process

import (
	"context"
	"encoding/base64"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/json"
)

// Base64InvalidDirection is returned when the Base64 processor is configured with an invalid direction setting.
const Base64InvalidDirection = errors.Error("Base64InvalidDirection")

// Base64InvalidAlphabet is returned when the Base64 processor is configured with an invalid alphabet setting.
const Base64InvalidAlphabet = errors.Error("Base64InvalidAlphabet")

/*
Base64Options contains custom options for the Base64 processor:
	direction:
		the direction of the encoding, either to (encode) or from (decode) base64
	alphabet (optional, defaults to "std"):
		the base64 alphabet to use, either std (https://www.rfc-editor.org/rfc/rfc4648.html#section-4) or url (https://www.rfc-editor.org/rfc/rfc4648.html#section-5)
*/
type Base64Options struct {
	Direction string `mapstructure:"direction"`
	Alphabet  string `mapstructure:"alphabet"`
}

// Base64 implements the Byter and Channeler interfaces and converts bytes to and from Base64. More information is available in the README.
type Base64 struct {
	Condition condition.OperatorConfig `mapstructure:"condition"`
	Input     Input                    `mapstructure:"input"`
	Output    Output                   `mapstructure:"output"`
	Options   Base64Options            `mapstructure:"options"`
}

// Channel processes a data channel of byte slices with the Base64 processor. Conditions are optionally applied on the channel data to enable processing.
func (p Base64) Channel(ctx context.Context, ch <-chan []byte) (<-chan []byte, error) {
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

// Byte processes a byte slice with the Base64 processor.
func (p Base64) Byte(ctx context.Context, data []byte) ([]byte, error) {
	if p.Options.Alphabet == "" {
		p.Options.Alphabet = "std"
	}

	// JSON object processing
	if p.Input.Key != "" && p.Output.Key != "" {
		value := json.Get(data, p.Input.Key)

		if !value.IsArray() {
			tmp := []byte(value.String())
			if p.Options.Direction == "from" {
				result, err := fromBase64(tmp, p.Options.Alphabet)
				if err != nil {
					return nil, err
				}
				return json.Set(data, p.Output.Key, result)
			} else if p.Options.Direction == "to" {
				result, err := toBase64(tmp, p.Options.Alphabet)
				if err != nil {
					return nil, err
				}
				return json.Set(data, p.Output.Key, result)
			} else {
				return nil, Base64InvalidDirection
			}
		}

		var array []string
		for _, v := range value.Array() {
			tmp := []byte(v.String())
			if p.Options.Direction == "from" {
				result, err := fromBase64(tmp, p.Options.Alphabet)
				if err != nil {
					return nil, err
				}
				array = append(array, string(result))
			} else if p.Options.Direction == "to" {
				result, err := toBase64(tmp, p.Options.Alphabet)
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
	tmp := data
	if p.Input.Key != "" {
		// convert string to bytes for base64 conversion
		v := json.Get(data, p.Input.Key)
		tmp = []byte(v.String())
	}

	var result []byte
	var err error
	if p.Options.Direction == "from" {
		result, err = fromBase64(tmp, p.Options.Alphabet)
		if err != nil {
			return nil, err
		}
	} else if p.Options.Direction == "to" {
		result, err = toBase64(tmp, p.Options.Alphabet)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, Base64InvalidDirection
	}

	if p.Output.Key != "" {
		return json.Set(data, p.Output.Key, result)
	}

	return result, nil
}

func fromBase64(data []byte, alphabet string) ([]byte, error) {
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

func toBase64(data []byte, alphabet string) ([]byte, error) {
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
