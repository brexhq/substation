package process

import (
	"context"
	"unicode/utf8"

	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/base64"
	"github.com/brexhq/substation/internal/errors"
)

// Base64JSONDecodedBinary is returned when the Base64 processor is configured to decode output to JSON, but the output contains binary data and cannot be written as valid JSON.
const Base64JSONDecodedBinary = errors.Error("Base64JSONDecodedBinary")

/*
Base64 processes data by converting it to and from base64 encoding. The processor supports these patterns:
	JSON:
	  	{"base64":"Zm9v"} >>> {"base64":"foo"}
	data:
		Zm9v >>> foo

When loaded with a factory, the processor uses this JSON configuration:
	{
		"type": "base64",
		"settings": {
			"options": {
				"direction": "from"
			},
			"input_key": "base64",
			"output_key": "base64"
		}
	}
*/
type Base64 struct {
	Options   Base64Options    `json:"options"`
	Condition condition.Config `json:"condition"`
	InputKey  string           `json:"input_key"`
	OutputKey string           `json:"output_key"`
}

/*
Base64Options contains custom options for the Base64 processor:
	Direction:
		the direction of the encoding
		must be one of:
			to: encode to base64
			from: decode from base64
*/
type Base64Options struct {
	Direction string `json:"direction"`
}

// ApplyBatch processes a slice of encapsulated data with the Base64 processor. Conditions are optionally applied to the data to enable processing.
func (p Base64) ApplyBatch(ctx context.Context, caps []config.Capsule) ([]config.Capsule, error) {
	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, fmt.Errorf("process base64 applybatch: %v", err)
	}

	caps, err = conditionallyApplyBatch(ctx, caps, op, p)
	if err != nil {
		return nil, fmt.Errorf("process base64 applybatch: %v", err)
	}

	return caps, nil
}

// Apply processes encapsulated data with the Base64 processor.
func (p Base64) Apply(ctx context.Context, cap config.Capsule) (config.Capsule, error) {
	// error early if required options are missing
	if p.Options.Direction == "" {
		return cap, fmt.Errorf("process base64 apply: options %+v: %v", p.Options, ProcessorMissingRequiredOptions)
	}

	// JSON processing
	if p.InputKey != "" && p.OutputKey != "" {
		result := cap.Get(p.InputKey).String()
		tmp := []byte(result)

		var value []byte
		switch p.Options.Direction {
		case "from":
			decode, err := base64.Decode(tmp)
			if err != nil {
				return cap, fmt.Errorf("process base64 apply: %v", err)
			}

			if !utf8.Valid(decode) {
				return cap, fmt.Errorf("process base64 apply: %v", Base64JSONDecodedBinary)
			}

			value = decode
		case "to":
			value = base64.Encode(tmp)
		default:
			return cap, fmt.Errorf("process base64 apply: direction %s: %v", p.Options.Direction, ProcessorInvalidDirection)
		}

		if err := cap.Set(p.OutputKey, value); err != nil {
			return cap, fmt.Errorf("process base64 apply: %v", err)
		}

		return cap, nil
	}

	// data processing
	if p.InputKey == "" && p.OutputKey == "" {
		var value []byte
		switch p.Options.Direction {
		case "from":
			decode, err := base64.Decode(cap.GetData())
			if err != nil {
				return cap, fmt.Errorf("process base64 apply: %v", err)
			}

			value = decode
		case "to":
			value = base64.Encode(cap.GetData())
		default:
			return cap, fmt.Errorf("process base64 apply: direction %s: %v", p.Options.Direction, ProcessorInvalidDirection)
		}

		cap.SetData(value)
		return cap, nil
	}

	return cap, fmt.Errorf("process base64 apply: inputkey %s outputkey %s: %v", p.InputKey, p.OutputKey, ProcessorInvalidDataPattern)
}
