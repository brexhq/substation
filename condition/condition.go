package condition

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
)

// errOperatorMissingInspectors is returned when an Operator that requires
// inspectors is created with no inspectors.
var errOperatorMissingInspectors = fmt.Errorf("missing inspectors")

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
func NewInspector(ctx context.Context, cfg config.Config) (Inspector, error) {
	switch cfg.Type {
	case "bitmath":
		return newInspBitmath(ctx, cfg)
	case "condition":
		return newInspCondition(ctx, cfg)
	case "content":
		return newInspContent(ctx, cfg)
	case "for_each":
		return newInspForEach(ctx, cfg)
	case "ip":
		return newInspIP(ctx, cfg)
	case "json_schema":
		return newInspJSONSchema(ctx, cfg)
	case "json_valid":
		return newInspJSONValid(ctx, cfg)
	case "length":
		return newInspLength(ctx, cfg)
	case "random":
		return newInspRandom(ctx, cfg)
	case "regexp":
		return newInspRegExp(ctx, cfg)
	case "strings":
		return newInspStrings(ctx, cfg)
	default:
		return nil, fmt.Errorf("condition: new_inspector: type %q settings %+v: %v", cfg.Type, cfg.Settings, errors.ErrInvalidFactoryInput)
	}
}

// NewInspectors accepts one or more Inspector configurations and returns configured inspectors.
func NewInspectors(ctx context.Context, cfg ...config.Config) ([]Inspector, error) {
	var inspectors []Inspector
	for _, c := range cfg {
		Inspector, err := NewInspector(ctx, c)
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
func NewOperator(ctx context.Context, cfg Config) (Operator, error) {
	inspectors, err := NewInspectors(ctx, cfg.Inspectors...)
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

// condition evaluates data with a condition (operator and inspectors).
//
// This inspector supports the object handling patterns of the inspectors passed to the condition.
type inspCondition struct {
	condition
	Options Config `json:"options"`

	op Operator
}

// Creates a new condition inspector.
func newInspCondition(ctx context.Context, cfg config.Config) (c inspCondition, err error) {
	if err = config.Decode(cfg.Settings, &c); err != nil {
		return inspCondition{}, err
	}

	c.op, err = NewOperator(ctx, c.Options)
	if err != nil {
		return inspCondition{}, err
	}

	return c, nil
}

func (c inspCondition) String() string {
	return toString(c)
}

// Inspect evaluates encapsulated data with the condition inspector.
func (c inspCondition) Inspect(ctx context.Context, capsule config.Capsule) (output bool, err error) {
	// this inspector does not directly interpret data, instead the
	// capsule is passed through and each configured inspector
	// applies its own data interpretation.
	matched, err := c.op.Operate(ctx, capsule)
	if err != nil {
		return false, err
	}

	if c.Negate {
		return !matched, nil
	}

	return matched, nil
}
