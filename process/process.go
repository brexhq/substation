package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
)

// errInvalidDataPattern is returned when a processor is configured with an invalid data access pattern. This is commonly caused by improperly set input and output settings.
var errInvalidDataPattern = errors.Error("invalid data access pattern")

// errInvalidDirection is returned when a processor is configured with an invalid direction setting.
const errInvalidDirection = errors.Error("invalid direction")

// errMissingRequiredOptions is returned when a processor does not have the required options to properly execute.
const errMissingRequiredOptions = errors.Error("missing required options")

// errInvalidFactoryInput is returned when an unsupported processor is referenced in any Factory.
const errInvalidFactoryInput = errors.Error("invalid factory input")

// Applicator is an interface for applying a processor to encapsulated data.
type Applicator interface {
	Apply(context.Context, config.Capsule) (config.Capsule, error)
}

// Apply accepts one or many Applicators and applies processors in series to encapsulated data.
func Apply(ctx context.Context, capsule config.Capsule, apps ...Applicator) (config.Capsule, error) {
	var err error

	for _, app := range apps {
		capsule, err = app.Apply(ctx, capsule)
		if err != nil {
			return capsule, err
		}
	}

	return capsule, nil
}

// ApplyByte is a convenience function for applying one or many Applicators to bytes.
func ApplyByte(ctx context.Context, data []byte, apps ...Applicator) ([]byte, error) {
	capsule := config.NewCapsule()
	capsule.SetData(data)

	newCap, err := Apply(ctx, capsule, apps...)
	if err != nil {
		return nil, err
	}

	return newCap.Data(), nil
}

// MakeApplicators accepts multiple processor configs and returns populated Applicators. This is a convenience function for generating many Applicators.
func MakeApplicators(cfg []config.Config) ([]Applicator, error) {
	var apps []Applicator

	for _, c := range cfg {
		app, err := ApplicatorFactory(c)
		if err != nil {
			return nil, err
		}
		apps = append(apps, app)
	}

	return apps, nil
}

// ApplicatorFactory returns a configured Applicator from a config. This is the recommended method for retrieving ready-to-use Applicators.
func ApplicatorFactory(cfg config.Config) (Applicator, error) {
	switch t := cfg.Type; t {
	case "base64":
		var p Base64
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "capture":
		var p Capture
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "case":
		var p Case
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "concat":
		var p Concat
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "convert":
		var p Convert
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "copy":
		var p Copy
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "delete":
		var p Delete
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "domain":
		var p Domain
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "dynamodb":
		var p DynamoDB
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "flatten":
		var p Flatten
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "for_each":
		var p ForEach
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "group":
		var p Group
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "gzip":
		var p Gzip
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "hash":
		var p Hash
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "insert":
		var p Insert
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "lambda":
		var p Lambda
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "math":
		var p Math
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "pipeline":
		var p Pipeline
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "pretty_print":
		var p PrettyPrint
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "replace":
		var p Replace
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "split":
		var p Split
		_ = config.Decode(cfg.Settings, &p)

		return p, nil
	case "time":
		var p Time
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	default:
		return nil, fmt.Errorf("process settings %+v: %v", cfg.Settings, errInvalidFactoryInput)
	}
}

// BatchApplicator is an interface for applying a processor to a slice of encapsulated data.
type BatchApplicator interface {
	ApplyBatch(context.Context, []config.Capsule) ([]config.Capsule, error)
}

// ApplyBatch accepts one or many BatchApplicators and applies processors in series to a slice of encapsulated data.
func ApplyBatch(ctx context.Context, batch []config.Capsule, apps ...BatchApplicator) ([]config.Capsule, error) {
	var err error

	for _, app := range apps {
		batch, err = app.ApplyBatch(ctx, batch)
		if err != nil {
			return nil, err
		}
	}

	return batch, nil
}

// MakeBatchApplicators accepts multiple processor configs and returns populated BatchApplicators. This is a convenience function for generating many BatchApplicators.
func MakeBatchApplicators(cfg []config.Config) ([]BatchApplicator, error) {
	var apps []BatchApplicator

	for _, c := range cfg {
		app, err := BatchApplicatorFactory(c)
		if err != nil {
			return nil, err
		}

		apps = append(apps, app)
	}

	return apps, nil
}

// BatchApplicatorFactory returns a configured BatchApplicator from a config. This is the recommended method for retrieving ready-to-use BatchApplicators.
func BatchApplicatorFactory(cfg config.Config) (BatchApplicator, error) {
	switch t := cfg.Type; t {
	case "aggregate":
		var p Aggregate
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "base64":
		var p Base64
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "capture":
		var p Capture
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "case":
		var p Case
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "concat":
		var p Concat
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "convert":
		var p Convert
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "copy":
		var p Copy
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "count":
		var p Count
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "delete":
		var p Delete
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "domain":
		var p Domain
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "drop":
		var p Drop
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "dynamodb":
		var p DynamoDB
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "expand":
		var p Expand
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "flatten":
		var p Flatten
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "for_each":
		var p ForEach
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "group":
		var p Group
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "gzip":
		var p Gzip
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "hash":
		var p Hash
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "insert":
		var p Insert
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "lambda":
		var p Lambda
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "math":
		var p Math
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "pipeline":
		var p Pipeline
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "pretty_print":
		var p PrettyPrint
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "replace":
		var p Replace
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "split":
		var p Split
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "time":
		var p Time
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	default:
		return nil, fmt.Errorf("process settings %+v: %v", cfg.Settings, errInvalidFactoryInput)
	}
}

// newBatch returns a Capsule slice with a minimum capacity of 10. This is used to speed up batch processing.
func newBatch(s *[]config.Capsule) []config.Capsule {
	if len(*s) > 10 {
		return make([]config.Capsule, 0, len(*s))
	}
	return make([]config.Capsule, 0, 10)
}

// conditionallyApplyBatch uses conditions to dynamically apply processors to a slice of encapsulated data. This is a convenience function for the ApplyBatch method used in most processors.
func conditionallyApplyBatch(ctx context.Context, capsules []config.Capsule, op condition.Operator, apps ...Applicator) ([]config.Capsule, error) {
	newCaps := newBatch(&capsules)

	for _, capsule := range capsules {
		ok, err := op.Operate(ctx, capsule)
		if err != nil {
			return nil, err
		}

		if !ok {
			newCaps = append(newCaps, capsule)
			continue
		}

		capsule, err := Apply(ctx, capsule, apps...)
		if err != nil {
			return nil, err
		}

		newCaps = append(newCaps, capsule)
	}

	return newCaps, nil
}
