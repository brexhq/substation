package condition

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
	mess "github.com/brexhq/substation/message"
)

// errOperatorMissingInspectors is returned when an Operator that requires
// inspectors is created with no inspectors.
var errOperatorMissingInspectors = fmt.Errorf("missing inspectors")

type Inspector interface {
	Inspect(context.Context, *mess.Message) (bool, error)
}

// NewInspector returns a configured Inspector from an Inspector configuration.
func NewInspector(ctx context.Context, cfg config.Config) (Inspector, error) {
	switch cfg.Type {
	case "insp_content":
		return newInspContent(ctx, cfg)
	case "insp_ip":
		return newInspIP(ctx, cfg)
	case "insp_json_valid":
		return newInspJSONValid(ctx, cfg)
	case "insp_length":
		return newInspLength(ctx, cfg)
	case "insp_random":
		return newInspRandom(ctx, cfg)
	case "insp_regexp":
		return newInspRegExp(ctx, cfg)
	case "insp_strings":
		return newInspStrings(ctx, cfg)
	case "meta_condition":
		return newMetaInspCondition(ctx, cfg)
	case "meta_for_each":
		return newMetaInspForEach(ctx, cfg)
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
	message, err := mess.New(
		mess.SetData(data),
	)
	if err != nil {
		return false, err
	}

	return inspect.Inspect(ctx, message)
}

type Operator interface {
	Operate(context.Context, *mess.Message) (bool, error)
}

// NewOperator returns a configured Operator from an Operator configuration.
func NewOperator(ctx context.Context, cfg Config) (Operator, error) {
	inspectors, err := NewInspectors(ctx, cfg.Inspectors...)
	if err != nil {
		return nil, err
	}

	switch cfg.Operator {
	case "all":
		return &opAll{inspectors}, nil
	case "any":
		return &opAny{inspectors}, nil
	case "none":
		return &opNone{inspectors}, nil
	default:
		return &opEmpty{}, nil
	}
}

// OperateBytes is a convenience function for applying an Operator to bytes.
func OperateBytes(ctx context.Context, data []byte, op Operator) (bool, error) {
	message, err := mess.New(
		mess.SetData(data),
	)
	if err != nil {
		return false, err
	}

	return op.Operate(ctx, message)
}

type opAll struct {
	Inspectors []Inspector `json:"inspectors"`
}

func (o *opAll) String() string {
	b, _ := json.Marshal(o)
	return string(b)
}

// Operate returns true if all inspectors return true, otherwise it returns false.
func (o *opAll) Operate(ctx context.Context, message *mess.Message) (bool, error) {
	if message.IsControl() {
		return false, nil
	}

	if len(o.Inspectors) == 0 {
		return false, fmt.Errorf("condition: operate: inspectors %+v: %v", o, errOperatorMissingInspectors)
	}

	for _, i := range o.Inspectors {
		ok, err := i.Inspect(ctx, message)
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

func (o *opAny) String() string {
	b, _ := json.Marshal(o)
	return string(b)
}

// Operate returns true if any inspectors return true, otherwise it returns false.
func (o *opAny) Operate(ctx context.Context, message *mess.Message) (bool, error) {
	if message.IsControl() {
		return false, nil
	}

	if len(o.Inspectors) == 0 {
		return false, fmt.Errorf("condition: operate: inspectors %+v: %v", o, errOperatorMissingInspectors)
	}

	for _, i := range o.Inspectors {
		ok, err := i.Inspect(ctx, message)
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

func (o *opNone) String() string {
	b, _ := json.Marshal(o)
	return string(b)
}

// Operate returns true if all inspectors return false, otherwise it returns true.
func (o *opNone) Operate(ctx context.Context, message *mess.Message) (bool, error) {
	if message.IsControl() {
		return false, nil
	}

	if len(o.Inspectors) == 0 {
		return false, fmt.Errorf("condition: operate: inspectors %+v: %v", o, errOperatorMissingInspectors)
	}

	for _, i := range o.Inspectors {
		ok, err := i.Inspect(ctx, message)
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

func (o *opEmpty) String() string {
	b, _ := json.Marshal(o)
	return string(b)
}

// Operate always returns true.
func (o *opEmpty) Operate(ctx context.Context, message *mess.Message) (bool, error) {
	return true, nil
}

// Config is used with NewOperator to produce new operators.
type Config struct {
	Operator   string          `json:"operator"`
	Inspectors []config.Config `json:"inspectors"`
}
