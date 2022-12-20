package process

import (
	"context"
	"fmt"
	"unicode/utf8"

	"github.com/brexhq/substation/config"
	ibase64 "github.com/brexhq/substation/internal/base64"
	"github.com/brexhq/substation/internal/errors"
)

// errBase64DecodedBinary is returned when the Base64 processor is configured
// to decode output to JSON, but the output contains binary data and cannot be
// written as valid JSON.
const errBase64DecodedBinary = errors.Error("cannot write binary as JSON")

// base64 processes data by converting it to and from base64.
//
// This processor supports the data and object handling patterns.
type _base64 struct {
	process
	Options _base64Options `json:"options"`
}

type _base64Options struct {
	// Direction determines whether data is encoded or decoded.
	//
	// Must be one of:
	//
	// - to: encode to base64
	//
	// - from: decode from base64
	Direction string `json:"direction"`
}

// Close closes resources opened by the processor.
func (p _base64) Close(context.Context) error {
	return nil
}

// Batch processes one or more capsules with the processor. Conditions are
// optionally applied to the data to enable processing.
func (p _base64) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	return conditionalApply(ctx, capsules, p.Condition, p)
}

// Apply processes a capsule with the processor.
func (p _base64) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	// error early if required options are missing
	if p.Options.Direction == "" {
		return capsule, fmt.Errorf("process base64: options %+v: %v", p.Options, errMissingRequiredOptions)
	}

	// JSON processing
	if p.Key != "" && p.SetKey != "" {
		result := capsule.Get(p.Key).String()
		tmp := []byte(result)

		var value []byte
		switch p.Options.Direction {
		case "from":
			decode, err := ibase64.Decode(tmp)
			if err != nil {
				return capsule, fmt.Errorf("process base64: %v", err)
			}

			if !utf8.Valid(decode) {
				return capsule, fmt.Errorf("process base64: %v", errBase64DecodedBinary)
			}

			value = decode
		case "to":
			value = ibase64.Encode(tmp)
		default:
			return capsule, fmt.Errorf("process base64: direction %s: %v", p.Options.Direction, errInvalidDirection)
		}

		if err := capsule.Set(p.SetKey, value); err != nil {
			return capsule, fmt.Errorf("process base64: %v", err)
		}

		return capsule, nil
	}

	// data processing
	if p.Key == "" && p.SetKey == "" {
		var value []byte
		switch p.Options.Direction {
		case "from":
			decode, err := ibase64.Decode(capsule.Data())
			if err != nil {
				return capsule, fmt.Errorf("process base64: %v", err)
			}

			value = decode
		case "to":
			value = ibase64.Encode(capsule.Data())
		default:
			return capsule, fmt.Errorf("process base64: direction %s: %v", p.Options.Direction, errInvalidDirection)
		}

		capsule.SetData(value)
		return capsule, nil
	}

	return capsule, fmt.Errorf("process base64: Key %s SetKey %s: %v", p.Key, p.SetKey, errInvalidDataPattern)
}
