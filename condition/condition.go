package condition

import (
	"fmt"

	"github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
)

// InspectorInvalidFactoryConfig is used when an unsupported Inspector is referenced in InspectorFactory
const InspectorInvalidFactoryConfig = errors.Error("InspectorInvalidFactoryConfig")

// OperatorInvalidFactoryConfig is used when an unsupported Operator is referenced in OperatorFactory
const OperatorInvalidFactoryConfig = errors.Error("OperatorInvalidFactoryConfig")

// OperatorMissingInspectors is used when an Operator that requres Inspectors is created with no inspectors
const OperatorMissingInspectors = errors.Error("OperatorMissingInspectors")

// Inspector is the interface shared by all inspector methods.
type Inspector interface {
	Inspect([]byte) (bool, error)
}

// Operator is the interface shared by all operator methods. Most operators contain a list of Inspectors that the operand applies to.
type Operator interface {
	Operate([]byte) (bool, error)
}

// AND implements the Operator interface and applies the boolean AND logic to configured inspectors.
type AND struct {
	Inspectors []Inspector
}

// Operate returns true if all Inspectors return true, otherwise it returns false.
func (o AND) Operate(data []byte) (bool, error) {
	if len(o.Inspectors) == 0 {
		return false, OperatorMissingInspectors
	}

	for _, i := range o.Inspectors {
		ok, err := i.Inspect(data)

		if err != nil {
			return false, err
		}
		// return false if any Check fails
		if !ok {
			return false, nil
		}
	}

	// return tue if all Checks pass
	return true, nil
}

// OR implements the Operator interface and applies the boolean OR logic to configured inspectors.
type OR struct {
	Inspectors []Inspector
}

// Operate returns true if any Inspectors return true, otherwise it returns false.
func (o OR) Operate(data []byte) (bool, error) {
	if len(o.Inspectors) == 0 {
		return false, OperatorMissingInspectors
	}

	for _, i := range o.Inspectors {
		ok, err := i.Inspect(data)
		if err != nil {
			return false, err
		}
		// return true if any Check passes
		if ok {
			return true, nil
		}
	}

	// return false if all Checks fail
	return false, nil
}

// NAND implements the Operator interface and applies the boolean NAND logic to configured inspectors.
type NAND struct {
	Inspectors []Inspector
}

// Operate returns true if all Inspectors return false, otherwise it returns true.
func (o NAND) Operate(data []byte) (bool, error) {
	if len(o.Inspectors) == 0 {
		return false, OperatorMissingInspectors
	}

	for _, i := range o.Inspectors {
		ok, err := i.Inspect(data)
		if err != nil {
			return false, err
		}
		// return true if any Check fails
		if !ok {
			return true, nil
		}
	}

	// return false if all Checks pass
	return false, nil
}

// NOR implements the Operator interface and applies the boolean NOR logic to configured inspectors.
type NOR struct {
	Inspectors []Inspector
}

// Operate returns true if any Inspectors return false, otherwise it returns true.
func (o NOR) Operate(data []byte) (bool, error) {
	if len(o.Inspectors) == 0 {
		return false, OperatorMissingInspectors
	}

	for _, i := range o.Inspectors {
		ok, err := i.Inspect(data)
		if err != nil {
			return false, err
		}
		// return false if any Check passes
		if ok {
			return false, nil
		}
	}

	// return true if all Checks fail
	return true, nil
}

// Default implements the Operator interface.
type Default struct{}

// Operate always returns true. This operator cannot be called directly and is chosen from the OperatorFactory if no valid operator is provided.
func (o Default) Operate(data []byte) (bool, error) {
	return true, nil
}

// OperatorConfig contains an array of InspectorConfig that are used to evaluate data.
type OperatorConfig struct {
	Operator   string
	Inspectors []config.Config
}

// OperatorFactory loads Operators from an OperatorConfig. This is the recommended function for retrieving ready-to-use Operators.
func OperatorFactory(cfg OperatorConfig) (Operator, error) {
	inspectors, err := MakeInspectors(cfg.Inspectors)
	if err != nil {
		return nil, err
	}

	switch op := cfg.Operator; op {
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

// InspectorFactory loads Inspectors from an InspectorConfig. This is the recommended function for retrieving ready-to-use Inspectors.
func InspectorFactory(cfg config.Config) (Inspector, error) {
	switch t := cfg.Type; t {
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
	case "regexp":
		var i RegExp
		config.Decode(cfg.Settings, &i)
		return i, nil
	case "strings":
		var i Strings
		config.Decode(cfg.Settings, &i)
		return i, nil
	default:
		return nil, fmt.Errorf("err retrieving %s from factory: %v", t, InspectorInvalidFactoryConfig)
	}
}

// MakeInspectors is a convenience function for making several Inspectors.
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
