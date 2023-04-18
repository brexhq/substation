package process

import (
	"context"
	"crypto/md5"
	"crypto/sha256"
	"fmt"

	"golang.org/x/exp/slices"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
)

// errHashInvalidAlgorithm is returned when the hash processor is configured with an invalid algorithm.
const errHashInvalidAlgorithm = errors.Error("invalid algorithm")

// hash processes data by calculating hashes (https://en.wikipedia.org/wiki/CryptographicprocHash_function).
//
// This processor supports the data and object handling patterns.
type procHash struct {
	process
	Options procHashOptions `json:"options"`
}

type procHashOptions struct {
	// Algorithm is the hashing algorithm applied to the data.
	//
	// Must be one of:
	//
	// - md5
	//
	// - sha256
	Algorithm string `json:"algorithm"`
}

// Create a new hash processor.
func newProcHash(ctx context.Context, cfg config.Config) (p procHash, err error) {
	if err = config.Decode(cfg.Settings, &p); err != nil {
		return procHash{}, err
	}

	p.operator, err = condition.NewOperator(ctx, p.Condition)
	if err != nil {
		return procHash{}, err
	}

	//  validate option.algorithm
	if !slices.Contains(
		[]string{
			"md5",
			"sha256",
		},
		p.Options.Algorithm) {
		return procHash{}, fmt.Errorf("process: hash: algorithm %q: %v", p.Options.Algorithm, errors.ErrInvalidOption)
	}

	return p, nil
}

// String returns the processor settings as an object.
func (p procHash) String() string {
	return toString(p)
}

// Closes resources opened by the processor.
func (p procHash) Close(context.Context) error {
	return nil
}

// Batch processes one or more capsules with the processor. Conditions are
// optionally applied to the data to enable processing.
func (p procHash) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	return batchApply(ctx, capsules, p, p.operator)
}

// Apply processes a capsule with the processor.
func (p procHash) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
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
			return capsule, fmt.Errorf("process: hash: algorithm %s: %v", p.Options.Algorithm, errHashInvalidAlgorithm)
		}

		if err := capsule.Set(p.SetKey, value); err != nil {
			return capsule, fmt.Errorf("process: hash: %v", err)
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
			return capsule, fmt.Errorf("process: hash: algorithm %s: %v", p.Options.Algorithm, errHashInvalidAlgorithm)
		}

		capsule.SetData([]byte(value))
		return capsule, nil
	}

	return capsule, fmt.Errorf("process: hash: key %s set_key %s: %v", p.Key, p.SetKey, errInvalidDataPattern)
}
