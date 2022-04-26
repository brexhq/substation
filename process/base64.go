package process

import (
	"context"
	"encoding/base64"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/internal/errors"
)

// Base64InvalidDirection is used when an invalid direction setting is given to the processor
const Base64InvalidDirection = errors.Error("Base64InvalidDirection")

// Base64InvalidAlphabet is used when an invalid alphabet setting is given to the processor
const Base64InvalidAlphabet = errors.Error("Base64InvalidAlphabet")

/*
Base64Options contain custom options settings for this processor.

Direction: the direction of the encoding, either to (encode) or from (decode) base64.
Alphabet (optional): the base64 alphabet to use, either std (https://www.rfc-editor.org/rfc/rfc4648.html#section-4) or url (https://www.rfc-editor.org/rfc/rfc4648.html#section-5); defaults to std.
*/
type Base64Options struct {
	Direction string `mapstructure:"direction"`
	Alphabet  string `mapstructure:"alphabet"`
}

// Base64 implements the Byter and Channeler interfaces and converts bytes to and from Base64. More information is available in the README.
type Base64 struct {
	Condition condition.OperatorConfig `mapstructure:"condition"`
	Options   Base64Options            `mapstructure:"options"`
}

// Channel processes a data channel of bytes with this processor. Conditions can be optionally applied on the channel data to enable processing.
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

// Byte processes a byte slice with this processor.
func (p Base64) Byte(ctx context.Context, data []byte) ([]byte, error) {
	if p.Options.Alphabet == "" {
		p.Options.Alphabet = "std"
	}

	if p.Options.Direction == "from" {
		res, err := fromBase64(data, p.Options.Alphabet)
		if err != nil {
			return nil, err
		}

		return res, nil
	} else if p.Options.Direction == "to" {
		res, err := toBase64(data, p.Options.Alphabet)
		if err != nil {
			return nil, err
		}

		return res, nil
	} else {
		return nil, Base64InvalidDirection
	}
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
