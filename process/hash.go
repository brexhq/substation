package process

import (
	"context"
	"crypto/md5"
	"crypto/sha256"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
)

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

/*
Hash processes encapsulated data by calculating hashes. The processor supports these patterns:
	JSON:
		{"hash":"foo"} >>> {"hash":"acbd18db4cc2f85cedef654fccc4a4d8"}
	data:
		foo >>> acbd18db4cc2f85cedef654fccc4a4d8

The processor uses this Jsonnet configuration:
	{
		type: 'hash',
		settings: {
			options: {
				algorithm: 'md5',
			},
			input_key: 'hash',
			output_key: 'hash',
		},
	}
*/
type Hash struct {
	Options   HashOptions              `json:"options"`
	Condition condition.OperatorConfig `json:"condition"`
	InputKey  string                   `json:"input_key"`
	OutputKey string                   `json:"output_key"`
}

// ApplyBatch processes a slice of encapsulated data with the Hash processor. Conditions are optionally applied to the data to enable processing.
func (p Hash) ApplyBatch(ctx context.Context, caps []config.Capsule) ([]config.Capsule, error) {
	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, fmt.Errorf("applybatch settings %+v: %w", p, err)
	}

	caps, err = conditionallyApplyBatch(ctx, caps, op, p)
	if err != nil {
		return nil, fmt.Errorf("applybatch settings %+v: %w", p, err)
	}

	return caps, nil
}

// Apply processes encapsulated data with the Hash processor.
func (p Hash) Apply(ctx context.Context, cap config.Capsule) (config.Capsule, error) {
	// error early if required options are missing
	if p.Options.Algorithm == "" {
		return cap, fmt.Errorf("applicator settings %+v: %w", p, ProcessorInvalidSettings)
	}

	// JSON processing
	if p.InputKey != "" && p.OutputKey != "" {
		result := cap.Get(p.InputKey).String()
		switch p.Options.Algorithm {
		case "md5":
			sum := md5.Sum([]byte(result))
			cap.Set(p.OutputKey, fmt.Sprintf("%x", sum))
		case "sha256":
			sum := sha256.Sum256([]byte(result))
			cap.Set(p.OutputKey, fmt.Sprintf("%x", sum))
		}

		return cap, nil
	}

	// data processing
	if p.InputKey == "" && p.OutputKey == "" {
		switch p.Options.Algorithm {
		case "md5":
			sum := md5.Sum(cap.GetData())
			sf := fmt.Sprintf("%x", sum)
			cap.SetData([]byte(sf))
		case "sha256":
			sum := sha256.Sum256(cap.GetData())
			sf := fmt.Sprintf("%x", sum)
			cap.SetData([]byte(sf))
		}

		return cap, nil
	}

	return cap, fmt.Errorf("applicator settings %+v: %w", p, ProcessorInvalidSettings)
}
