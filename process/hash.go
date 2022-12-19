package process

import (
	"context"
	"crypto/md5"
	"crypto/sha256"
	"fmt"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
)

// errhashInvalidAlgorithm is returned when the hash processor is configured with an invalid algorithm.
const errhashInvalidAlgorithm = errors.Error("invalid algorithm")

type hash struct {
	process
	Options hashOptions `json:"options"`
}

type hashOptions struct {
	Algorithm string `json:"algorithm"`
}

// Close closes resources opened by the hash processor.
func (p hash) Close(context.Context) error {
	return nil
}

func (p hash) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	capsules, err := conditionalApply(ctx, capsules, p.Condition, p)
	if err != nil {
		return nil, fmt.Errorf("process hash: %v", err)
	}

	return capsules, nil
}

// Apply processes encapsulated data with the hash processor.
func (p hash) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
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
			return capsule, fmt.Errorf("process hash: algorithm %s: %v", p.Options.Algorithm, errhashInvalidAlgorithm)
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
			return capsule, fmt.Errorf("process hash: algorithm %s: %v", p.Options.Algorithm, errhashInvalidAlgorithm)
		}

		capsule.SetData([]byte(value))
		return capsule, nil
	}

	return capsule, fmt.Errorf("process hash: inputkey %s outputkey %s: %v", p.Key, p.SetKey, errInvalidDataPattern)
}
