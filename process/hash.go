package process

import (
	"context"
	"crypto/md5"
	"crypto/sha256"
	"fmt"
	"math/rand"
	"time"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
)

// errHashInvalidAlgorithm is returned when the Hash processor is configured with an invalid algorithm.
const errHashInvalidAlgorithm = errors.Error("invalid algorithm")

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
		hashing algorithm applied to the data
		must be one of:
			md5
			sha256
*/
type HashOptions struct {
	Algorithm string `json:"algorithm"`
}

// ApplyBatch processes a slice of encapsulated data with the Hash processor. Conditions are optionally applied to the data to enable processing.
func (p Hash) ApplyBatch(ctx context.Context, capsules []config.Capsule) ([]config.Capsule, error) {
	rand.Seed(time.Now().UnixNano())
	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, fmt.Errorf("process hash: %v", err)
	}

	capsules, err = conditionallyApplyBatch(ctx, capsules, op, p)
	if err != nil {
		return nil, fmt.Errorf("process hash: %v", err)
	}

	return capsules, nil
}

// Apply processes encapsulated data with the Hash processor.
func (p Hash) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	// error early if required options are missing
	if p.Options.Algorithm == "" {
		return capsule, fmt.Errorf("process hash: options %+v: %v", p.Options, errMissingRequiredOptions)
	}

	// JSON processing
	if p.InputKey != "" && p.OutputKey != "" {
		result := capsule.Get(p.InputKey).String()

		var value string
		switch p.Options.Algorithm {
		case "md5":
			sum := md5.Sum([]byte(result))
			value = fmt.Sprintf("%x", sum)
		case "sha256":
			sum := sha256.Sum256([]byte(result))
			value = fmt.Sprintf("%x", sum)
		default:
			return capsule, fmt.Errorf("process hash: algorithm %s: %v", p.Options.Algorithm, errHashInvalidAlgorithm)
		}

		if err := capsule.Set(p.OutputKey, value); err != nil {
			return capsule, fmt.Errorf("process hash: %v", err)
		}

		return capsule, nil
	}

	// data processing
	if p.InputKey == "" && p.OutputKey == "" {
		var value string
		switch p.Options.Algorithm {
		case "md5":
			sum := md5.Sum(capsule.Data())
			value = fmt.Sprintf("%x", sum)
		case "sha256":
			sum := sha256.Sum256(capsule.Data())
			value = fmt.Sprintf("%x", sum)
		default:
			return capsule, fmt.Errorf("process hash: algorithm %s: %v", p.Options.Algorithm, errHashInvalidAlgorithm)
		}

		capsule.SetData([]byte(value))
		return capsule, nil
	}

	return capsule, fmt.Errorf("process hash: inputkey %s outputkey %s: %v", p.InputKey, p.OutputKey, errInvalidDataPattern)
}
