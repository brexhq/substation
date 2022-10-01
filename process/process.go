package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
)

// processorInvalidDataPattern is returned when a processor is configured with an invalid data access pattern. This is commonly caused by improperly set input and output settings.
const processorInvalidDataPattern = errors.Error("ProcessorIncorrectDataSettings")

// processorInvalidDirection is returned when a processor is configured with an invalid direction setting.
const processorInvalidDirection = errors.Error("processorInvalidDirection")

// processorMissingRequiredOptions is returned when a processor does not have the required options to properly execute.
const processorMissingRequiredOptions = errors.Error("processorMissingRequiredOptions")

// applyInvalidFactoryConfig is returned when an unsupported Task processor is referenced in Factory.
const applyInvalidFactoryConfig = errors.Error("applyInvalidFactoryConfig")

// applyBatchInvalidFactoryConfig is returned when an unsupported Batch processor is referenced in BatchFactory.
const applyBatchInvalidFactoryConfig = errors.Error("applyBatchInvalidFactoryConfig")

// Applicator is an interface for applying a processor to encapsulated data.
type Applicator interface {
	Apply(context.Context, config.Capsule) (config.Capsule, error)
}

// Apply accepts one or many Applicators and applies processors in series to encapsulated data.
func Apply(ctx context.Context, cap config.Capsule, apps ...Applicator) (config.Capsule, error) {
	var err error

	for _, app := range apps {
		cap, err = app.Apply(ctx, cap)
		if err != nil {
			return cap, err
		}
	}

	return cap, nil
}

// ApplyByte is a convenience function for applying one or many Applicators to bytes.
func ApplyByte(ctx context.Context, data []byte, apps ...Applicator) ([]byte, error) {
	cap := config.NewCapsule()
	cap.SetData(data)

	newCap, err := Apply(ctx, cap, apps...)
	if err != nil {
		return nil, err
	}

	return newCap.GetData(), nil
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
		return nil, fmt.Errorf("process settings %+v: %v", cfg.Settings, applyInvalidFactoryConfig)
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
		return nil, fmt.Errorf("process settings %+v: %v", cfg.Settings, applyBatchInvalidFactoryConfig)
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
func conditionallyApplyBatch(ctx context.Context, caps []config.Capsule, op condition.Operator, apps ...Applicator) ([]config.Capsule, error) {
	newCaps := newBatch(&caps)

	for _, cap := range caps {
		ok, err := op.Operate(ctx, cap)
		if err != nil {
			return nil, err
		}

		if !ok {
			newCaps = append(newCaps, cap)
			continue
		}

		cap, err := Apply(ctx, cap, apps...)
		if err != nil {
			return nil, err
		}

		newCaps = append(newCaps, cap)
	}

	return newCaps, nil
}
