package condition

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
)

// errInvalidFactoryInput is returned when an unsupported inspector is referenced in inspectorFactory.
const errInvalidFactoryInput = errors.Error("invalid factory input")

// errOperatorMissinginspectors is returned when an Operator that requires inspectors is created with no inspectors.
const errOperatorMissinginspectors = errors.Error("missing inspectors")

type condition struct {
	// Key retrieves a value from an object for inspection.
	//
	// This is optional for inspectors that support inspecting non-object data.
	Key string `json:"key"`
	// Negate reverses the outcome of an inspection (true becomes false and false becomes true).
	//
	// This is optional and defaults to false.
	Negate bool `json:"negate"`
}

// inspector is the interface shared by all inspector methods.
type inspector interface {
	Inspect(context.Context, config.Capsule) (bool, error)
}

// InspectByte is a convenience function for applying an inspector to bytes.
func InspectBytes(ctx context.Context, data []byte, inspect inspector) (bool, error) {
	capsule := config.NewCapsule()
	capsule.SetData(data)

	return inspect.Inspect(ctx, capsule)
}

// MakeInspectors accepts multiple inspector configs and returns populated inspectors. This is a convenience function for generating many inspectors.
func MakeInspectors(cfg []config.Config) ([]inspector, error) {
	var inspectors []inspector
	for _, c := range cfg {
		inspector, err := InspectorFactory(c)
		if err != nil {
			return nil, err
		}
		inspectors = append(inspectors, inspector)
	}

	return inspectors, nil
}

// inspectorFactory returns a configured inspector from a config. This is the recommended method for retrieving ready-to-use inspectors.
func InspectorFactory(cfg config.Config) (inspector, error) {
	switch cfg.Type {
	case "content":
		var i content
		_ = config.Decode(cfg.Settings, &i)
		return i, nil
	case "for_each":
		var i forEach
		_ = config.Decode(cfg.Settings, &i)
		return i, nil
	case "ip":
		var i ip
		_ = config.Decode(cfg.Settings, &i)
		return i, nil
	case "json_schema":
		var i jsonSchema
		_ = config.Decode(cfg.Settings, &i)
		return i, nil
	case "json_valid":
		var i jsonValid
		_ = config.Decode(cfg.Settings, &i)
		return i, nil
	case "length":
		var i length
		_ = config.Decode(cfg.Settings, &i)
		return i, nil
	case "random":
		var i random
		_ = config.Decode(cfg.Settings, &i)
		return i, nil
	case "regexp":
		var i regExp
		_ = config.Decode(cfg.Settings, &i)
		return i, nil
	case "strings":
		var i strings
		_ = config.Decode(cfg.Settings, &i)
		return i, nil
	default:
		return nil, fmt.Errorf("condition inspectorfactory: type %q, settings %+v: %v", cfg.Type, cfg.Settings, errInvalidFactoryInput)
	}
}

// operator is the interface shared by all operator methods. Operators apply a series of inspectors to and verify the state (aka "condition") of data.
type operator interface {
	Operate(context.Context, config.Capsule) (bool, error)
}

// and implements the Operator interface and applies the boolean AND logic to configured inspectors.
type and struct {
	inspectors []inspector
}

// Operate returns true if all inspectors return true, otherwise it returns false.
func (o and) Operate(ctx context.Context, capsule config.Capsule) (bool, error) {
	if len(o.inspectors) == 0 {
		return false, fmt.Errorf("condition operate: inspectors %+v: %v", o, errOperatorMissinginspectors)
	}

	for _, i := range o.inspectors {
		ok, err := i.Inspect(ctx, capsule)
		if err != nil {
			return false, err
		}

		// return false if any check fails
		if !ok {
			return false, nil
		}
	}

	// return tue if all checks pass
	return true, nil
}

// or implements the Operator interface and applies the boolean or logic to configured inspectors.
type or struct {
	inspectors []inspector
}

// Operate returns true if any inspectors return true, otherwise it returns false.
func (o or) Operate(ctx context.Context, capsule config.Capsule) (bool, error) {
	if len(o.inspectors) == 0 {
		return false, fmt.Errorf("condition operate: inspectors %+v: %v", o, errOperatorMissinginspectors)
	}

	for _, i := range o.inspectors {
		ok, err := i.Inspect(ctx, capsule)
		if err != nil {
			return false, err
		}

		// return true if any check passes
		if ok {
			return true, nil
		}
	}

	// return false if all checks fail
	return false, nil
}

// nand implements the Operator interface and applies the boolean nand logic to configured inspectors.
type nand struct {
	inspectors []inspector
}

// Operate returns true if all inspectors return false, otherwise it returns true.
func (o nand) Operate(ctx context.Context, capsule config.Capsule) (bool, error) {
	if len(o.inspectors) == 0 {
		return false, fmt.Errorf("condition operate: inspectors %+v: %v", o, errOperatorMissinginspectors)
	}

	for _, i := range o.inspectors {
		ok, err := i.Inspect(ctx, capsule)
		if err != nil {
			return false, err
		}

		// return true if any check fails
		if !ok {
			return true, nil
		}
	}

	// return false if all checks pass
	return false, nil
}

// nor implements the Operator interface and applies the boolean nor logic to configured inspectors.
type nor struct {
	inspectors []inspector
}

// Operate returns true if any inspectors return false, otherwise it returns true.
func (o nor) Operate(ctx context.Context, capsule config.Capsule) (bool, error) {
	if len(o.inspectors) == 0 {
		return false, fmt.Errorf("condition operate: inspectors %+v: %v", o, errOperatorMissinginspectors)
	}

	for _, i := range o.inspectors {
		ok, err := i.Inspect(ctx, capsule)
		if err != nil {
			return false, err
		}

		// return false if any check passes
		if ok {
			return false, nil
		}
	}

	// return true if all checks fail
	return true, nil
}

// empty implements the Operator interface.
type empty struct{}

// empty always returns true. This is the default operator returned by  OperatorFactory.
func (o empty) Operate(ctx context.Context, capsule config.Capsule) (bool, error) {
	return true, nil
}

// OperatorFactory returns a configured Operator from a config. This is the recommended method for retrieving ready-to-use Operators.
func OperatorFactory(cfg Config) (operator, error) {
	inspectors, err := MakeInspectors(cfg.Inspectors)
	if err != nil {
		return nil, err
	}

	switch cfg.Operator {
	case "and":
		return and{inspectors}, nil
	case "nand":
		return nand{inspectors}, nil
	case "or":
		return or{inspectors}, nil
	case "nor":
		return nor{inspectors}, nil
	default:
		return empty{}, nil
	}
}

// Config is used with OperatorFactory to produce new Operators from JSON configurations.
type Config struct {
	Operator   string
	Inspectors []config.Config
}
