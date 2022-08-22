package process

import (
	"context"
	"crypto/md5"
	"crypto/sha256"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
)

// HashInvalidAlgorithm is returned when the Hash processor is configured with an invalid algorithm.
const HashInvalidAlgorithm = errors.Error("HashInvalidAlgorithm")

/*
Hash processes data by calculating hashes. The processor supports these patterns:
	JSON:
		{"hash":"foo"} >>> {"hash":"acbd18db4cc2f85cedef654fccc4a4d8"}
	data:
		foo >>> acbd18db4cc2f85cedef654fccc4a4d8

When loaded with a factory, the processor uses this JSON configuration:
	{
		"type": "hash",
		"settings": {
			"options": {
				"algorithm": "md5"
			},
			"input_key": "hash",
			"output_key": "hash"
		}
	}
*/
type Hash struct {
	Options   HashOptions      `json:"options"`
	Condition condition.Config `json:"condition"`
	InputKey  string           `json:"input_key"`
	OutputKey string           `json:"output_key"`
}

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

// ApplyBatch processes a slice of encapsulated data with the Hash processor. Conditions are optionally applied to the data to enable processing.
func (p Hash) ApplyBatch(ctx context.Context, caps []config.Capsule) ([]config.Capsule, error) {
	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, fmt.Errorf("applybatch settings %+v: %v", p, err)
	}

	caps, err = conditionallyApplyBatch(ctx, caps, op, p)
	if err != nil {
		return nil, fmt.Errorf("applybatch settings %+v: %v", p, err)
	}

	return caps, nil
}

// Apply processes encapsulated data with the Hash processor.
func (p Hash) Apply(ctx context.Context, cap config.Capsule) (config.Capsule, error) {
	// error early if required options are missing
	if p.Options.Algorithm == "" {
		return cap, fmt.Errorf("apply settings %+v: %w", p, ProcessorInvalidSettings)
	}

	// JSON processing
	if p.InputKey != "" && p.OutputKey != "" {
		result := cap.Get(p.InputKey).String()

		var value string
		switch p.Options.Algorithm {
		case "md5":
			sum := md5.Sum([]byte(result))
			value = fmt.Sprintf("%x", sum)
		case "sha256":
			sum := sha256.Sum256([]byte(result))
			value = fmt.Sprintf("%x", sum)
		default:
			return cap, fmt.Errorf("apply settings %+v: %w", p, HashInvalidAlgorithm)
		}

		if err := cap.Set(p.OutputKey, value); err != nil {
			return cap, fmt.Errorf("apply settings %+v: %v", p, err)
		}

		return cap, nil
	}

	// data processing
	if p.InputKey == "" && p.OutputKey == "" {
		var value string
		switch p.Options.Algorithm {
		case "md5":
			sum := md5.Sum(cap.GetData())
			value = fmt.Sprintf("%x", sum)
		case "sha256":
			sum := sha256.Sum256(cap.GetData())
			value = fmt.Sprintf("%x", sum)
		default:
			return cap, fmt.Errorf("apply settings %+v: %w", p, HashInvalidAlgorithm)
		}

		cap.SetData([]byte(value))
		return cap, nil
	}

	return cap, fmt.Errorf("apply settings %+v: %w", p, ProcessorInvalidSettings)
}
