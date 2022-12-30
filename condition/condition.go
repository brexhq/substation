package condition

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
)

// errInvalidFactoryInput is returned when an unsupported type is
// referenced in any factory.
const errInvalidFactoryInput = errors.Error("invalid factory input")

// errOperatorMissinginspectors is returned when an operator that requires
// inspectors is created with no inspectors.
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

func toString(i interface{}) string {
	switch v := i.(type) {
	case inspector:
		b, _ := json.Marshal(v)
		return string(b)
	case operator:
		b, _ := json.Marshal(v)
		return string(b)
	default:
		return ""
	}
}

type inspector interface {
	Inspect(context.Context, config.Capsule) (bool, error)
}

// MakeInspector returns a configured inspector from an inspector configuration.
func MakeInspector(cfg config.Config) (inspector, error) {
	switch cfg.Type {
	case "content":
		var i _content
		_ = config.Decode(cfg.Settings, &i)
		return i, nil
	case "for_each":
		var i _forEach
		_ = config.Decode(cfg.Settings, &i)
		return i, nil
	case "ip":
		var i _ip
		_ = config.Decode(cfg.Settings, &i)
		return i, nil
	case "json_schema":
		var i _jsonSchema
		_ = config.Decode(cfg.Settings, &i)
		return i, nil
	case "json_valid":
		var i _jsonValid
		_ = config.Decode(cfg.Settings, &i)
		return i, nil
	case "length":
		var i _length
		_ = config.Decode(cfg.Settings, &i)
		return i, nil
	case "random":
		var i _random
		_ = config.Decode(cfg.Settings, &i)
		return i, nil
	case "regexp":
		var i _regExp
		_ = config.Decode(cfg.Settings, &i)
		return i, nil
	case "strings":
		var i _strings
		_ = config.Decode(cfg.Settings, &i)
		return i, nil
	default:
		return nil, fmt.Errorf("condition: make_inspector: type %q settings %+v: %v", cfg.Type, cfg.Settings, errInvalidFactoryInput)
	}
}

// MakeInspectors accepts one or more inspector configurations and returns configured inspectors.
func MakeInspectors(cfg ...config.Config) ([]inspector, error) {
	var inspectors []inspector
	for _, c := range cfg {
		inspector, err := MakeInspector(c)
		if err != nil {
			return nil, err
		}
		inspectors = append(inspectors, inspector)
	}

	return inspectors, nil
}

// InspectByte is a convenience function for applying an inspector to bytes.
func InspectBytes(ctx context.Context, data []byte, inspect inspector) (bool, error) {
	capsule := config.NewCapsule()
	capsule.SetData(data)

	return inspect.Inspect(ctx, capsule)
}

type operator interface {
	Operate(context.Context, config.Capsule) (bool, error)
}

// MakeOperator returns a configured operator from an operator configuration.
func MakeOperator(cfg Config) (operator, error) {
	inspectors, err := MakeInspectors(cfg.Inspectors...)
	if err != nil {
		return nil, err
	}

	switch cfg.Operator {
	case "all":
		return _all{inspectors}, nil
	case "any":
		return _any{inspectors}, nil
	case "none":
		return _none{inspectors}, nil
	default:
		return _empty{}, nil
	}
}

// OperateBytes is a convenience function for applying an operator to bytes.
func OperateBytes(ctx context.Context, data []byte, op operator) (bool, error) {
	capsule := config.NewCapsule()
	capsule.SetData(data)

	return op.Operate(ctx, capsule)
}

type _all struct {
	Inspectors []inspector `json:"inspectors"`
}

func (o _all) String() string {
	return toString(o)
}

// Operate returns true if all inspectors return true, otherwise it returns false.
func (o _all) Operate(ctx context.Context, capsule config.Capsule) (bool, error) {
	if len(o.Inspectors) == 0 {
		return false, fmt.Errorf("condition: operate: inspectors %+v: %v", o, errOperatorMissinginspectors)
	}

	for _, i := range o.Inspectors {
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

type _any struct {
	Inspectors []inspector `json:"inspectors"`
}

func (o _any) String() string {
	return toString(o)
}

// Operate returns true if any inspectors return true, otherwise it returns false.
func (o _any) Operate(ctx context.Context, capsule config.Capsule) (bool, error) {
	if len(o.Inspectors) == 0 {
		return false, fmt.Errorf("condition: operate: inspectors %+v: %v", o, errOperatorMissinginspectors)
	}

	for _, i := range o.Inspectors {
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

// none implements the operator interface.
type _none struct {
	Inspectors []inspector `json:"inspectors"`
}

func (o _none) String() string {
	return toString(o)
}

// Operate returns true if all inspectors return false, otherwise it returns true.
func (o _none) Operate(ctx context.Context, capsule config.Capsule) (bool, error) {
	if len(o.Inspectors) == 0 {
		return false, fmt.Errorf("condition: operate: inspectors %+v: %v", o, errOperatorMissinginspectors)
	}

	for _, i := range o.Inspectors {
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

type _empty struct{}

func (o _empty) String() string {
	return toString(o)
}

// Operate always returns true.
func (o _empty) Operate(ctx context.Context, capsule config.Capsule) (bool, error) {
	return true, nil
}

// Config is used with MakeOperator to produce new operators.
type Config struct {
	Operator   string          `json:"operator"`
	Inspectors []config.Config `json:"inspectors"`
}
