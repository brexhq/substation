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
	case Applier:
		b, _ := json.Marshal(v)
		return string(b)
	case Batcher:
		b, _ := json.Marshal(v)
		return string(b)
	default:
		return ""
	}
}

type Applier interface {
	Apply(context.Context, config.Capsule) (config.Capsule, error)
	Close(context.Context) error
}

// NewApplier returns a configured Applier from a processor configuration.
func NewApplier(cfg config.Config) (Applier, error) {
	switch cfg.Type {
	case "aws_dynamodb":
		var p procAWSDynamoDB
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "aws_lambda":
		var p procAWSLambda
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "base64":
		var p procBase64
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "capture":
		var p procCapture
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "case":
		var p procCase
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "convert":
		var p procConvert
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "copy":
		var p procCopy
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "delete":
		var p procDelete
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "dns":
		var p procDNS
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "domain":
		var p procDomain
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "flatten":
		var p procFlatten
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "for_each":
		var p procForEach
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "group":
		var p procGroup
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "gzip":
		var p procGzip
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "hash":
		var p procHash
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "insert":
		var p procInsert
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "ip_database":
		var p procIPDatabase
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "join":
		var p procJoin
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "kv_store":
		var p procKVStore
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "math":
		var p procMath
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "pipeline":
		var p procPipeline
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "pretty_print":
		var p procPrettyPrint
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "replace":
		var p procReplace
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "split":
		var p procSplit
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "time":
		var p procTime
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	default:
		return nil, fmt.Errorf("process: new_applier: type %q settings %+v: %v", cfg.Type, cfg.Settings, errors.ErrInvalidFactoryInput)
	}
}

// NewAppliers accepts one or more processor configurations and returns configured appliers.
func NewAppliers(cfg ...config.Config) ([]Applier, error) {
	var apps []Applier

	for _, c := range cfg {
		a, err := NewApplier(c)
		if err != nil {
			return nil, err
		}
		apps = append(apps, a)
	}

	return apps, nil
}

// CloseAppliers closes all appliers and returns an error if any close fails.
func CloseAppliers(ctx context.Context, appliers ...Applier) error {
	for _, a := range appliers {
		if err := a.Close(ctx); err != nil {
			return err
		}
	}

	return nil
}

// Apply applies processors in series to encapsulated data.
func Apply(ctx context.Context, capsule config.Capsule, appliers ...Applier) (config.Capsule, error) {
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
func ApplyBytes(ctx context.Context, data []byte, appliers ...Applier) ([]byte, error) {
	capsule := config.NewCapsule()
	capsule.SetData(data)

	newCapsule, err := Apply(ctx, capsule, appliers...)
	if err != nil {
		return nil, err
	}

	return newCapsule.Data(), nil
}

type Batcher interface {
	Batch(context.Context, ...config.Capsule) ([]config.Capsule, error)
	Close(context.Context) error
}

// NewBatcher returns a configured Batcher from a processor configuration.
func NewBatcher(cfg config.Config) (Batcher, error) { //nolint: cyclop // ignore cyclomatic complexity
	switch cfg.Type {
	case "aggregate":
		var p procAggregate
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "aws_dynamodb":
		var p procAWSDynamoDB
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "aws_lambda":
		var p procAWSLambda
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "base64":
		var p procBase64
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "capture":
		var p procCapture
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "case":
		var p procCase
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "convert":
		var p procConvert
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "copy":
		var p procCopy
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "count":
		var p procCount
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "delete":
		var p procDelete
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "dns":
		var p procDNS
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "domain":
		var p procDomain
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "drop":
		var p procDrop
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "expand":
		var p procExpand
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "flatten":
		var p procFlatten
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "for_each":
		var p procForEach
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "group":
		var p procGroup
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "gzip":
		var p procGzip
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "hash":
		var p procHash
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "insert":
		var p procInsert
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "ip_database":
		var p procIPDatabase
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "join":
		var p procJoin
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "kv_store":
		var p procKVStore
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "math":
		var p procMath
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "pipeline":
		var p procPipeline
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "pretty_print":
		var p procPrettyPrint
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "replace":
		var p procReplace
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "split":
		var p procSplit
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	case "time":
		var p procTime
		_ = config.Decode(cfg.Settings, &p)
		return p, nil
	default:
		return nil, fmt.Errorf("process: new_batcher: type %q settings %+v: %v", cfg.Type, cfg.Settings, errors.ErrInvalidFactoryInput)
	}
}

// NewBatchers accepts one or more processor configurations and returns configured batchers.
func NewBatchers(cfg ...config.Config) ([]Batcher, error) {
	var bats []Batcher

	for _, c := range cfg {
		b, err := NewBatcher(c)
		if err != nil {
			return nil, err
		}

		bats = append(bats, b)
	}

	return bats, nil
}

// CloseBatchers closes all batchers and returns an error if any close fails.
func CloseBatchers(ctx context.Context, batchers ...Batcher) error {
	for _, b := range batchers {
		if err := b.Close(ctx); err != nil {
			return err
		}
	}

	return nil
}

// Batch accepts one or more batchers and applies processors in series to encapsulated data.
func Batch(ctx context.Context, batch []config.Capsule, batchers ...Batcher) ([]config.Capsule, error) {
	var err error

	for _, Batcher := range batchers {
		batch, err = Batcher.Batch(ctx, batch...)
		if err != nil {
			return nil, err
		}
	}

	return batch, nil
}

// BatchBytes is a convenience function for applying processors in series to bytes.
func BatchBytes(ctx context.Context, data [][]byte, batchers ...Batcher) ([][]byte, error) {
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

func batchApply(ctx context.Context, capsules []config.Capsule, app Applier, c condition.Config) ([]config.Capsule, error) {
	op, err := condition.NewOperator(c)
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
