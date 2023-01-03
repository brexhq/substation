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

// errOperatorMissingInspectors is returned when an Operator that requires
// inspectors is created with no inspectors.
const errOperatorMissingInspectors = errors.Error("missing inspectors")

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
	case Inspector:
		b, _ := json.Marshal(v)
		return string(b)
	case Operator:
		b, _ := json.Marshal(v)
		return string(b)
	default:
		return ""
	}
}

type Inspector interface {
	Inspect(context.Context, config.Capsule) (bool, error)
}

// NewInspector returns a configured Inspector from an Inspector configuration.
func NewInspector(cfg config.Config) (Inspector, error) {
	switch cfg.Type {
	case "content":
		var i inspContent
		_ = config.Decode(cfg.Settings, &i)
		return i, nil
	case "for_each":
		var i inspForEach
		_ = config.Decode(cfg.Settings, &i)
		return i, nil
	case "ip":
		var i inspIP
		_ = config.Decode(cfg.Settings, &i)
		return i, nil
	case "json_schema":
		var i inspJSONSchema
		_ = config.Decode(cfg.Settings, &i)
		return i, nil
	case "json_valid":
		var i inspJSONValid
		_ = config.Decode(cfg.Settings, &i)
		return i, nil
	case "length":
		var i inspLength
		_ = config.Decode(cfg.Settings, &i)
		return i, nil
	case "random":
		var i inspRandom
		_ = config.Decode(cfg.Settings, &i)
		return i, nil
	case "regexp":
		var i inspRegExp
		_ = config.Decode(cfg.Settings, &i)
		return i, nil
	case "strings":
		var i inspStrings
		_ = config.Decode(cfg.Settings, &i)
		return i, nil
	default:
		return nil, fmt.Errorf("condition: make_inspector: type %q settings %+v: %v", cfg.Type, cfg.Settings, errInvalidFactoryInput)
	}
}

// NewInspectors accepts one or more Inspector configurations and returns configured inspectors.
func NewInspectors(cfg ...config.Config) ([]Inspector, error) {
	var inspectors []Inspector
	for _, c := range cfg {
		Inspector, err := NewInspector(c)
		if err != nil {
			return nil, err
		}
		inspectors = append(inspectors, Inspector)
	}

	return inspectors, nil
}

// InspectByte is a convenience function for applying an Inspector to bytes.
func InspectBytes(ctx context.Context, data []byte, inspect Inspector) (bool, error) {
	capsule := config.NewCapsule()
	capsule.SetData(data)

	return inspect.Inspect(ctx, capsule)
}

type Operator interface {
	Operate(context.Context, config.Capsule) (bool, error)
}

// NewOperator returns a configured Operator from an Operator configuration.
func NewOperator(cfg Config) (Operator, error) {
	inspectors, err := NewInspectors(cfg.Inspectors...)
	if err != nil {
		return nil, err
	}

	switch cfg.Operator {
	case "all":
		return opAll{inspectors}, nil
	case "any":
		return opAny{inspectors}, nil
	case "none":
		return opNone{inspectors}, nil
	default:
		return opEmpty{}, nil
	}
}

// OperateBytes is a convenience function for applying an Operator to bytes.
func OperateBytes(ctx context.Context, data []byte, op Operator) (bool, error) {
	capsule := config.NewCapsule()
	capsule.SetData(data)

	return op.Operate(ctx, capsule)
}

type opAll struct {
	Inspectors []Inspector `json:"inspectors"`
}

func (o opAll) String() string {
	return toString(o)
}

// Operate returns true if all inspectors return true, otherwise it returns false.
func (o opAll) Operate(ctx context.Context, capsule config.Capsule) (bool, error) {
	if len(o.Inspectors) == 0 {
		return false, fmt.Errorf("condition: operate: inspectors %+v: %v", o, errOperatorMissingInspectors)
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

type opAny struct {
	Inspectors []Inspector `json:"inspectors"`
}

func (o opAny) String() string {
	return toString(o)
}

// Operate returns true if any inspectors return true, otherwise it returns false.
func (o opAny) Operate(ctx context.Context, capsule config.Capsule) (bool, error) {
	if len(o.Inspectors) == 0 {
		return false, fmt.Errorf("condition: operate: inspectors %+v: %v", o, errOperatorMissingInspectors)
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

type opNone struct {
	Inspectors []Inspector `json:"inspectors"`
}

func (o opNone) String() string {
	return toString(o)
}

// Operate returns true if all inspectors return false, otherwise it returns true.
func (o opNone) Operate(ctx context.Context, capsule config.Capsule) (bool, error) {
	if len(o.Inspectors) == 0 {
		return false, fmt.Errorf("condition: operate: inspectors %+v: %v", o, errOperatorMissingInspectors)
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

type opEmpty struct{}

func (o opEmpty) String() string {
	return toString(o)
}

// Operate always returns true.
func (o opEmpty) Operate(ctx context.Context, capsule config.Capsule) (bool, error) {
	return true, nil
}

// Config is used with NewOperator to produce new operators.
type Config struct {
	Operator   string          `json:"operator"`
	Inspectors []config.Config `json:"inspectors"`
}
