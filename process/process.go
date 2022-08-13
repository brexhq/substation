package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
)

// ProcessorInvalidSettings is returned when a processor is configured with invalid settings. Common causes include improper input and output settings (e.g., missing keys) and missing required options.
const ProcessorInvalidSettings = errors.Error("ProcessorInvalidSettings")

// ApplyInvalidFactoryConfig is returned when an unsupported Task processor is referenced in Factory.
const ApplyInvalidFactoryConfig = errors.Error("ApplyInvalidFactoryConfig")

// ApplyBatchInvalidFactoryConfig is returned when an unsupported Batch processor is referenced in BatchFactory.
const ApplyBatchInvalidFactoryConfig = errors.Error("ApplyBatchInvalidFactoryConfig")

// Applicator is an interface for applying a processor to encapsulated data.
type Applicator interface {
	Apply(context.Context, config.Capsule) (config.Capsule, error)
}

// BatchApplicator is an interface for applying a processor to a slice of encapsulated data.
type BatchApplicator interface {
	ApplyBatch(context.Context, []config.Capsule) ([]config.Capsule, error)
}

// Apply accepts one or many Applicators and applies the processors in series to encapsulated data.
func Apply(ctx context.Context, cap config.Capsule, applicators ...Applicator) (config.Capsule, error) {
	var err error

	for _, app := range applicators {
		cap, err = app.Apply(ctx, cap)
		if err != nil {
			return cap, err
		}
	}

	return cap, nil
}

// ApplyBatch accepts one or many BatchApplicators and applies the processors in series to a slice of encapsulated data.
func ApplyBatch(ctx context.Context, batch []config.Capsule, applicators ...BatchApplicator) ([]config.Capsule, error) {
	var err error

	for _, app := range applicators {
		batch, err = app.ApplyBatch(ctx, batch)
		if err != nil {
			return nil, err
		}
	}

	return batch, nil
}

// Factory returns a configured Applicator from a config. This is the recommended method for retrieving ready-to-use Applicators.
func Factory(cfg config.Config) (Applicator, error) {
	switch t := cfg.Type; t {
	case "base64":
		var p Base64
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "capture":
		var p Capture
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "case":
		var p Case
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "concat":
		var p Concat
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "convert":
		var p Convert
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "copy":
		var p Copy
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "delete":
		var p Delete
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "domain":
		var p Domain
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "dynamodb":
		var p DynamoDB
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "flatten":
		var p Flatten
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "for_each":
		var p ForEach
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "group":
		var p Group
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "gzip":
		var p Gzip
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "hash":
		var p Hash
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "insert":
		var p Insert
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "lambda":
		var p Lambda
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "math":
		var p Math
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "pipeline":
		var p Pipeline
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "pretty_print":
		var p PrettyPrint
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "replace":
		var p Replace
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "split":
		var p Split
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "time":
		var p Time
		config.Decode(cfg.Settings, &p)
		return p, nil
	default:
		return nil, fmt.Errorf("process settings %+v: %w", cfg.Settings, ApplyInvalidFactoryConfig)
	}
}

// BatchFactory returns a configured BatchApplicator from a config. This is the recommended method for retrieving ready-to-use BatchApplicators.
func BatchFactory(cfg config.Config) (BatchApplicator, error) {
	switch t := cfg.Type; t {
	case "aggregate":
		var p Aggregate
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "base64":
		var p Base64
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "capture":
		var p Capture
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "case":
		var p Case
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "concat":
		var p Concat
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "convert":
		var p Convert
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "copy":
		var p Copy
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "count":
		var p Count
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "delete":
		var p Delete
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "domain":
		var p Domain
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "drop":
		var p Drop
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "dynamodb":
		var p DynamoDB
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "expand":
		var p Expand
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "flatten":
		var p Flatten
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "for_each":
		var p ForEach
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "group":
		var p Group
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "gzip":
		var p Gzip
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "hash":
		var p Hash
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "insert":
		var p Insert
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "lambda":
		var p Lambda
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "math":
		var p Math
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "pipeline":
		var p Pipeline
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "pretty_print":
		var p PrettyPrint
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "replace":
		var p Replace
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "split":
		var p Split
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "time":
		var p Time
		config.Decode(cfg.Settings, &p)
		return p, nil
	default:
		return nil, fmt.Errorf("process settings %+v: %w", cfg.Settings, ApplyBatchInvalidFactoryConfig)
	}
}

// MakeAll accepts multiple processor configs and returns populated Applicators. This is a convenience function for generating many Applicators.
func MakeAll(cfg []config.Config) ([]Applicator, error) {
	var applicators []Applicator

	for _, c := range cfg {
		applicator, err := Factory(c)
		if err != nil {
			return nil, err
		}
		applicators = append(applicators, applicator)
	}

	return applicators, nil
}

// MakeAllBatchApplicators accepts multiple processor configs and returns populated BatchApplicators. This is a convenience function for generating many BatchApplicators.
func MakeAllBatchApplicators(cfg []config.Config) ([]BatchApplicator, error) {
	var applicators []BatchApplicator

	for _, c := range cfg {
		applicator, err := BatchFactory(c)
		if err != nil {
			return nil, err
		}

		applicators = append(applicators, applicator)
	}

	return applicators, nil
}

// NewBatch returns a Capsule slice with a minimum capacity of 10. This is most frequently used to speed up batch processing.
func NewBatch(s *[]config.Capsule) []config.Capsule {
	if len(*s) > 10 {
		return make([]config.Capsule, 0, len(*s))
	}
	return make([]config.Capsule, 0, 10)
}

// Byte applies an Applicator to data.
func Byte(ctx context.Context, a Applicator, data []byte) ([]byte, error) {
	cap := config.NewCapsule()
	cap.SetData(data)

	newCap, err := a.Apply(ctx, cap)
	if err != nil {
		return nil, fmt.Errorf("byte settings %+v: %w", a, ProcessorInvalidSettings)
	}

	return newCap.GetData(), nil
}

// conditionallyApplyBatch uses conditions to dynamically apply processors to a slice of encapsulated data. This is a convenience function for the ApplyBatch method used in most processors.
func conditionallyApplyBatch(ctx context.Context, caps []config.Capsule, op condition.Operator, applicators ...Applicator) ([]config.Capsule, error) {
	slice := NewBatch(&caps)

	for _, cap := range caps {
		ok, err := op.Operate(cap)
		if err != nil {
			return nil, fmt.Errorf("filterbatch settings %+v: %w", op, err)
		}

		if !ok {
			slice = append(slice, cap)
			continue
		}

		cap, err := Apply(ctx, cap, applicators...)
		if err != nil {
			return nil, fmt.Errorf("filterbatch settings %+v: %w", op, err)
		}
		slice = append(slice, cap)
	}
	return slice, nil
}
