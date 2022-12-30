package process

import (
	"context"
	"encoding/json"
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

func toString(i interface{}) string {
	switch v := i.(type) {
	case applier:
		b, _ := json.Marshal(v)
		return string(b)
	case batcher:
		b, _ := json.Marshal(v)
		return string(b)
	default:
		return ""
	}
}

type applier interface {
	Apply(context.Context, config.Capsule) (config.Capsule, error)
	Close(context.Context) error
}

// MakeApplier returns a configured applier from a processor configuration.
func MakeApplier(cfg config.Config) (applier, error) {
	switch cfg.Type {
	case "aws_dynamodb":
		var p _awsDynamodb
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "aws_lambda":
		var p _awsLambda
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "base64":
		var p _base64
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "capture":
		var p _capture
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "case":
		var p _case
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "convert":
		var p _convert
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "copy":
		var p _copy
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "delete":
		var p _delete
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "dns":
		var p _dns
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "domain":
		var p _domain
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "flatten":
		var p _flatten
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "for_each":
		var p _forEach
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "group":
		var p _group
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "gzip":
		var p _gzip
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "hash":
		var p _hash
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "insert":
		var p _insert
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "ip_database":
		var p _ipDatabase
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "join":
		var p _join
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "math":
		var p _math
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "pipeline":
		var p _pipeline
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "pretty_print":
		var p _prettyPrint
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "replace":
		var p _replace
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "split":
		var p _split
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "time":
		var p _time
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	default:
		return nil, fmt.Errorf("process: make_applier: type %q settings %+v: %v", cfg.Type, cfg.Settings, errInvalidFactoryInput)
	}
}

// MakeAppliers accepts one or more processor configurations and returns configured appliers.
func MakeAppliers(cfg ...config.Config) ([]applier, error) {
	var apps []applier

	for _, c := range cfg {
		a, err := MakeApplier(c)
		if err != nil {
			return nil, err
		}
		apps = append(apps, a)
	}

	return apps, nil
}

// CloseAppliers closes all appliers and returns an error if any close fails.
func CloseAppliers(ctx context.Context, appliers ...applier) error {
	for _, a := range appliers {
		if err := a.Close(ctx); err != nil {
			return err
		}
	}

	return nil
}

// Apply applies processors in series to encapsulated data.
func Apply(ctx context.Context, capsule config.Capsule, appliers ...applier) (config.Capsule, error) {
	var err error

	for _, app := range appliers {
		capsule, err = app.Apply(ctx, capsule)
		if err != nil {
			return capsule, err
		}
	}

	return capsule, nil
}

// ApplyBytes is a convenience function for applying processors in series to bytes.
func ApplyBytes(ctx context.Context, data []byte, appliers ...applier) ([]byte, error) {
	capsule := config.NewCapsule()
	capsule.SetData(data)

	newCapsule, err := Apply(ctx, capsule, appliers...)
	if err != nil {
		return nil, err
	}

	return newCapsule.Data(), nil
}

type batcher interface {
	Batch(context.Context, ...config.Capsule) ([]config.Capsule, error)
	Close(context.Context) error
}

// MakeBatcher returns a configured batcher from a processor configuration.
func MakeBatcher(cfg config.Config) (batcher, error) {
	switch cfg.Type {
	case "aggregate":
		var p _aggregate
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "aws_dynamodb":
		var p _awsDynamodb
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "aws_lambda":
		var p _awsLambda
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "base64":
		var p _base64
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "capture":
		var p _capture
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "case":
		var p _case
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "convert":
		var p _convert
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "copy":
		var p _copy
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "count":
		var p _count
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "delete":
		var p _delete
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "dns":
		var p _dns
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "domain":
		var p _domain
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "drop":
		var p _drop
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "expand":
		var p _expand
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "flatten":
		var p _flatten
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "for_each":
		var p _forEach
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "group":
		var p _group
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "gzip":
		var p _gzip
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "hash":
		var p _hash
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "insert":
		var p _insert
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "ip_database":
		var p _ipDatabase
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "join":
		var p _join
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "math":
		var p _math
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "pipeline":
		var p _pipeline
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "pretty_print":
		var p _prettyPrint
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "replace":
		var p _replace
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "split":
		var p _split
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "time":
		var p _time
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	default:
		return nil, fmt.Errorf("process: make_batcher: type %q settings %+v: %v", cfg.Type, cfg.Settings, errInvalidFactoryInput)
	}
}

// MakeBatchers accepts one or more processor configurations and returns configured batchers.
func MakeBatchers(cfg ...config.Config) ([]batcher, error) {
	var bats []batcher

	for _, c := range cfg {
		b, err := MakeBatcher(c)
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

// newBatch returns a Capsule slice with a minimum capacity of 10. This is used to speed up batch processing.
func newBatch(s *[]config.Capsule) []config.Capsule {
	if len(*s) > 10 {
		return make([]config.Capsule, 0, len(*s))
	}
	return make([]config.Capsule, 0, 10)
}

func batchApply(ctx context.Context, capsules []config.Capsule, app applier, c condition.Config) ([]config.Capsule, error) {
	op, err := condition.MakeOperator(c)
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
