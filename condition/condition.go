package condition

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
)

// inspectorInvalidFactoryConfig is returned when an unsupported Inspector is referenced in InspectorFactory.
const inspectorInvalidFactoryConfig = errors.Error("inspectorInvalidFactoryConfig")

// operatorMissingInspectors is returned when an Operator that requres Inspectors is created with no inspectors.
const operatorMissingInspectors = errors.Error("operatorMissingInspectors")

// Inspector is the interface shared by all inspector methods.
type Inspector interface {
	Inspect(context.Context, config.Capsule) (bool, error)
}

// InspectByte is a convenience function for applying an Inspector to bytes.
func InspectByte(ctx context.Context, data []byte, inspect Inspector) (bool, error) {
	cap := config.NewCapsule()
	cap.SetData(data)

	return inspect.Inspect(ctx, cap)
}

// MakeInspectors accepts multiple inspector configs and returns populated Inspectors. This is a convenience function for generating many Inspectors.
func MakeInspectors(cfg []config.Config) ([]Inspector, error) {
	var inspectors []Inspector
	for _, c := range cfg {
		inspector, err := InspectorFactory(c)
		if err != nil {
			return nil, err
		}
		inspectors = append(inspectors, inspector)
	}

	return inspectors, nil
}

// InspectorFactory returns a configured Inspector from a config. This is the recommended method for retrieving ready-to-use Inspectors.
func InspectorFactory(cfg config.Config) (Inspector, error) {
	switch cfg.Type {
	case "content":
		var i Content
		config.Decode(cfg.Settings, &i)
		return i, nil
	case "ip":
		var i IP
		config.Decode(cfg.Settings, &i)
		return i, nil
	case "json_schema":
		var i JSONSchema
		config.Decode(cfg.Settings, &i)
		return i, nil
	case "json_valid":
		var i JSONValid
		config.Decode(cfg.Settings, &i)
		return i, nil
	case "length":
		var i Length
		config.Decode(cfg.Settings, &i)
		return i, nil
	case "random":
		var i Random
		config.Decode(cfg.Settings, &i)
		return i, nil
	case "regexp":
		var i RegExp
		config.Decode(cfg.Settings, &i)
		return i, nil
	case "strings":
		var i Strings
		config.Decode(cfg.Settings, &i)
		return i, nil
	default:
		return nil, fmt.Errorf("condition inspectorfactory: settings %+v: %v", cfg.Settings, inspectorInvalidFactoryConfig)
	}
}

// Operator is the interface shared by all operator methods. Operators apply a series of Inspectors to and verify the state (aka "condition") of data.
type Operator interface {
	Operate(context.Context, config.Capsule) (bool, error)
}

// AND implements the Operator interface and applies the boolean AND logic to configured inspectors.
type AND struct {
	Inspectors []Inspector
}

// Operate returns true if all Inspectors return true, otherwise it returns false.
func (o AND) Operate(ctx context.Context, cap config.Capsule) (bool, error) {
	if len(o.Inspectors) == 0 {
		return false, fmt.Errorf("condition operate: inspectors %+v: %v", o, operatorMissingInspectors)
	}

	for _, i := range o.Inspectors {
		ok, err := i.Inspect(ctx, cap)

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

// OR implements the Operator interface and applies the boolean OR logic to configured inspectors.
type OR struct {
	Inspectors []Inspector
}

// Operate returns true if any Inspectors return true, otherwise it returns false.
func (o OR) Operate(ctx context.Context, cap config.Capsule) (bool, error) {
	if len(o.Inspectors) == 0 {
		return false, fmt.Errorf("condition operate: inspectors %+v: %v", o, operatorMissingInspectors)
	}

	for _, i := range o.Inspectors {
		ok, err := i.Inspect(ctx, cap)
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

// NAND implements the Operator interface and applies the boolean NAND logic to configured inspectors.
type NAND struct {
	Inspectors []Inspector
}

// Operate returns true if all Inspectors return false, otherwise it returns true.
func (o NAND) Operate(ctx context.Context, cap config.Capsule) (bool, error) {
	if len(o.Inspectors) == 0 {
		return false, fmt.Errorf("condition operate: inspectors %+v: %v", o, operatorMissingInspectors)
	}

	for _, i := range o.Inspectors {
		ok, err := i.Inspect(ctx, cap)
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

// NOR implements the Operator interface and applies the boolean NOR logic to configured inspectors.
type NOR struct {
	Inspectors []Inspector
}

// Operate returns true if any Inspectors return false, otherwise it returns true.
func (o NOR) Operate(ctx context.Context, cap config.Capsule) (bool, error) {
	if len(o.Inspectors) == 0 {
		return false, fmt.Errorf("condition operate: inspectors %+v: %v", o, operatorMissingInspectors)
	}

	for _, i := range o.Inspectors {
		ok, err := i.Inspect(ctx, cap)
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

// Default implements the Operator interface.
type Default struct{}

// Operate always returns true. This is the default operator returned by  OperatorFactory.
func (o Default) Operate(ctx context.Context, cap config.Capsule) (bool, error) {
	return true, nil
}

// OperatorFactory returns a configured Operator from a config. This is the recommended method for retrieving ready-to-use Operators.
func OperatorFactory(cfg Config) (Operator, error) {
	inspectors, err := MakeInspectors(cfg.Inspectors)
	if err != nil {
		return nil, err
	}

	switch cfg.Operator {
	case "and":
		return AND{inspectors}, nil
	case "nand":
		return NAND{inspectors}, nil
	case "or":
		return OR{inspectors}, nil
	case "nor":
		return NOR{inspectors}, nil
	default:
		return Default{}, nil
	}
}

// Config is used with OperatorFactory to produce new Operators from JSON configurations.
type Config struct {
	Operator   string
	Inspectors []config.Config
}
