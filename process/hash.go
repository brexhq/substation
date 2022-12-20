package process

import (
	"context"
	"crypto/md5"
	"crypto/sha256"
	"fmt"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
)

// errHashInvalidAlgorithm is returned when the hash processor is configured with an invalid algorithm.
const errHashInvalidAlgorithm = errors.Error("invalid algorithm")

// hash processes data by calculating hashes (https://en.wikipedia.org/wiki/Cryptographic_hash_function).
//
// This processor supports the data and object handling patterns.
type _hash struct {
	process
	Options _hashOptions `json:"options"`
}

type _hashOptions struct {
	// Algorithm is the hashing algorithm applied to the data.
	//
	// Must be one of:
	//
	// - md5
	//
	// - sha256
	Algorithm string `json:"algorithm"`
}

// Close closes resources opened by the processor.
func (p _hash) Close(context.Context) error {
	return nil
}

// Batch processes one or more capsules with the processor. Conditions are
// optionally applied to the data to enable processing.
func (p _hash) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	capsules, err := conditionalApply(ctx, capsules, p.Condition, p)
	if err != nil {
		return nil, fmt.Errorf("process hash: %v", err)
	}

	return capsules, nil
}

// Apply processes a capsule with the processor.
func (p _hash) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	// error early if required options are missing
	if p.Options.Algorithm == "" {
		return capsule, fmt.Errorf("process hash: options %+v: %v", p.Options, errMissingRequiredOptions)
	}

	// JSON processing
	if p.Key != "" && p.SetKey != "" {
		result := capsule.Get(p.Key).String()

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

		if err := capsule.Set(p.SetKey, value); err != nil {
			return capsule, fmt.Errorf("process hash: %v", err)
		}

		return capsule, nil
	}

	// data processing
	if p.Key == "" && p.SetKey == "" {
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

	return capsule, fmt.Errorf("process hash: inputkey %s outputkey %s: %v", p.Key, p.SetKey, errInvalidDataPattern)
}
