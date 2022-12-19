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

type process struct {
	// Condition optionally enables processing depending on the outcome of data inspection.
	Condition condition.Config `json:"condition"`
	// Key retrieves a value from an object for processing.
	//
	// This is optional for processors that support processing non-object data.
	Key string `json:"key"`
	// SetKey inserts a processed value into an object.
	//
	// This is optional for processors that support processing non-object data.
	SetKey string `json:"set_key"`
	// IgnoreClose overrides attempts to close a processor.
	IgnoreClose bool `json:"ignore_close"`
	// IgnoreErrors overrides returning errors from a processor.
	IgnoreErrors bool `json:"ignore_errors"`
}

type applicator interface {
	Apply(context.Context, config.Capsule) (config.Capsule, error)
	Close(context.Context) error
}

// applicatorFactory returns a configured Applicator from a config. This is the recommended method for retrieving ready-to-use Applicators.
func applicatorFactory(cfg config.Config) (applicator, error) {
	switch cfg.Type {
	case "base64":
		var p base64
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "capture":
		var p capture
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "letter_case":
		var p letterCase
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "concat":
		var p concat
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "convert":
		var p convert
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "copy":
		var p copy
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "delete":
		var p delete
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "dns":
		var p dns
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "domain":
		var p domain
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "dynamodb":
		var p dynamodb
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "flatten":
		var p flatten
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "for_each":
		var p forEach
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "group":
		var p group
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "gzip":
		var p gzip
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "hash":
		var p hash
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "insert":
		var p insert
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "ip_database":
		var p ipDatabase
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "lambda":
		var p lambda
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "math":
		var p math
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "pipeline":
		var p pipeline
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "pretty_print":
		var p prettyPrint
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "replace":
		var p replace
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "split":
		var p split
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "time":
		var p time
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	default:
		return nil, fmt.Errorf("process settings %+v: %v", cfg.Settings, errInvalidFactoryInput)
	}
}

type batcher interface {
	Batch(context.Context, ...config.Capsule) ([]config.Capsule, error)
	Close(context.Context) error
}

func batcherFactory(cfg config.Config) (batcher, error) {
	switch cfg.Type {
	case "aggregate":
		var p aggregate
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "base64":
		var p base64
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "capture":
		var p capture
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "letter_case":
		var p letterCase
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "concat":
		var p concat
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "convert":
		var p convert
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "copy":
		var p copy
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "count":
		var p count
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "delete":
		var p delete
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "dns":
		var p dns
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "domain":
		var p domain
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "drop":
		var p drop
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "dynamodb":
		var p dynamodb
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "expand":
		var p expand
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "flatten":
		var p flatten
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "for_each":
		var p forEach
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "group":
		var p group
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "gzip":
		var p gzip
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "hash":
		var p hash
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "insert":
		var p insert
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "ip_database":
		var p ipDatabase
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "lambda":
		var p lambda
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "math":
		var p math
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "pipeline":
		var p pipeline
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "pretty_print":
		var p prettyPrint
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "replace":
		var p replace
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "split":
		var p split
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "time":
		var p time
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	default:
		return nil, fmt.Errorf("process settings %+v: %v", cfg.Settings, errInvalidFactoryInput)
	}
}

// Apply applies processors in series to encapsulated data.
func Apply(ctx context.Context, capsule config.Capsule, applicators ...applicator) (config.Capsule, error) {
	var err error

	for _, app := range applicators {
		capsule, err = app.Apply(ctx, capsule)
		if err != nil {
			return capsule, err
		}
	}

	return capsule, nil
}

// ApplyBytes is a convenience function for applying processors in series to bytes.
func ApplyBytes(ctx context.Context, data []byte, applicators ...applicator) ([]byte, error) {
	capsule := config.NewCapsule()
	capsule.SetData(data)

	newCapsule, err := Apply(ctx, capsule, applicators...)
	if err != nil {
		return nil, err
	}

	return newCapsule.Data(), nil
}

// MakeApplicators accepts one or more processor configurations and returns configured applicators.
func MakeApplicators(cfg ...config.Config) ([]applicator, error) {
	var apps []applicator

	for _, c := range cfg {
		a, err := applicatorFactory(c)
		if err != nil {
			return nil, err
		}
		apps = append(apps, a)
	}

	return apps, nil
}

// CloseApplicators closes all applicators and returns an error if any close fails.
func CloseApplicators(ctx context.Context, applicators ...applicator) error {
	for _, a := range applicators {
		if err := a.Close(ctx); err != nil {
			return err
		}
	}

	return nil
}

// Batch accepts one or more batchers and applies processors in series to encapsulated data.
func Batch(ctx context.Context, batch []config.Capsule, batchers ...batcher) ([]config.Capsule, error) {
	var err error

	for _, batcher := range batchers {
		batch, err = batcher.Batch(ctx, batch...)
		if err != nil {
			return nil, err
		}
	}

	return batch, nil
}

// BatchBytes is a convenience function for applying processors in series to bytes.
func BatchBytes(ctx context.Context, data [][]byte, batchers ...batcher) ([][]byte, error) {
	var capsules []config.Capsule
	capsule := config.NewCapsule()

	for _, d := range data {
		capsule.SetData(d)
		capsules = append(capsules, capsule)
	}

	batch, err := Batch(ctx, capsules, batchers...)
	if err != nil {
		return nil, err
	}

	var arr [][]byte
	for _, b := range batch {
		arr = append(arr, b.Data())
	}

	return arr, nil
}

// MakeBatchers accepts one or more processor configurations and returns populated batchers.
func MakeBatchers(cfg ...config.Config) ([]batcher, error) {
	var bats []batcher

	for _, c := range cfg {
		b, err := batcherFactory(c)
		if err != nil {
			return nil, err
		}

		bats = append(bats, b)
	}

	return bats, nil
}

// CloseBatchers closes all batchers and returns an error if any close fails.
func CloseBatchers(ctx context.Context, batchers ...batcher) error {
	for _, b := range batchers {
		if err := b.Close(ctx); err != nil {
			return err
		}
	}

	return nil
}

// newBatch returns a Capsule slice with a minimum capacity of 10. This is used to speed up batch processing.
func newBatch(s *[]config.Capsule) []config.Capsule {
	if len(*s) > 10 {
		return make([]config.Capsule, 0, len(*s))
	}
	return make([]config.Capsule, 0, 10)
}

func conditionalApply(ctx context.Context, capsules []config.Capsule, cond condition.Config, app applicator) ([]config.Capsule, error) {
	op, err := condition.OperatorFactory(cond)
	if err != nil {
		return nil, err
	}

	newCapsules := newBatch(&capsules)
	for _, c := range capsules {
		ok, err := op.Operate(ctx, c)
		if err != nil {
			return nil, err
		}

		if !ok {
			newCapsules = append(newCapsules, c)
			continue
		}

		newCapsule, err := app.Apply(ctx, c)
		if err != nil {
			return nil, err
		}

		newCapsules = append(newCapsules, newCapsule)
	}

	return newCapsules, nil
}

func batch(ctx context.Context, capsules []config.Capsule, apps ...applicator) ([]config.Capsule, error) {
	newCapsules := newBatch(&capsules)
	for _, c := range capsules {
		capsule, err := Apply(ctx, c, apps...)
		if err != nil {
			return nil, err
		}

		newCapsules = append(newCapsules, capsule)
	}

	return newCapsules, nil
}
