// Package condition provides functions for evaluating data.
package condition

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/message"
)

// errOperatorMissingInspectors is returned when an Operator that requires
// inspectors is created with no inspectors.
var errOperatorMissingInspectors = fmt.Errorf("missing inspectors")

type Config struct {
	Operator   string          `json:"operator"`
	Inspectors []config.Config `json:"inspectors"`
}

type inspector interface {
	Inspect(context.Context, *message.Message) (bool, error)
}

// newInspector returns a configured Inspector from an Inspector configuration.
func newInspector(ctx context.Context, cfg config.Config) (inspector, error) { //nolint: cyclop, gocyclo // ignore cyclomatic complexity
	switch cfg.Type {
	// Format inspectors.
	case "format_mime":
		return newFormatMIME(ctx, cfg)
	case "format_json":
		return newFormatJSON(ctx, cfg)
	// Meta inspectors.
	case "meta_condition":
		return newMetaCondition(ctx, cfg)
	case "meta_for_each":
		return newMetaForEach(ctx, cfg)
	case "meta_negate":
		return newMetaNegate(ctx, cfg)
	// Network inspectors.
	case "network_ip_global_unicast":
		return newNetworkIPGlobalUnicast(ctx, cfg)
	case "network_ip_link_local_multicast":
		return newNetworkIPLinkLocalMulticast(ctx, cfg)
	case "network_ip_link_local_unicast":
		return newNetworkIPLinkLocalUnicast(ctx, cfg)
	case "network_ip_loopback":
		return newNetworkIPLoopback(ctx, cfg)
	case "network_ip_multicast":
		return newNetworkIPMulticast(ctx, cfg)
	case "network_ip_private":
		return newNetworkIPPrivate(ctx, cfg)
	case "network_ip_unicast":
		return newNetworkIPUnicast(ctx, cfg)
	case "network_ip_unspecified":
		return newNetworkIPUnspecified(ctx, cfg)
	case "network_ip_valid":
		return newNetworkIPValid(ctx, cfg)
	// Number inspectors.
	case "number_greater_than":
		return newNumberGreaterThan(ctx, cfg)
	case "number_bitwise_and":
		return newNumberBitwiseAND(ctx, cfg)
	case "number_bitwise_or":
		return newNumberBitwiseOR(ctx, cfg)
	case "number_bitwise_xor":
		return newNumberBitwiseXOR(ctx, cfg)
	case "number_bitwise_not":
		return newNumberBitwiseNOT(ctx, cfg)
	case "number_length_less_than":
		return newNumberLengthLessThan(ctx, cfg)
	case "number_length_greater_than":
		return newNumberLengthGreaterThan(ctx, cfg)
	case "number_length_equal_to":
		return newNumberLengthEqualTo(ctx, cfg)
	// String inspectors.
	case "string_contains":
		return newStringContains(ctx, cfg)
	case "string_ends_with":
		return newStringEndsWith(ctx, cfg)
	case "string_equal_to":
		return newStringEqualTo(ctx, cfg)
	case "string_greater_than":
		return newStringGreaterThan(ctx, cfg)
	case "string_less_than":
		return newStringLessThan(ctx, cfg)
	case "string_starts_with":
		return newStringStartsWith(ctx, cfg)
	case "string_match":
		return newStringMatch(ctx, cfg)
	// Utility inspectors.
	case "utility_random":
		return newUtilityRandom(ctx, cfg)
	default:
		return nil, fmt.Errorf("condition: new_inspector: type %q settings %+v: %v", cfg.Type, cfg.Settings, errors.ErrInvalidFactoryInput)
	}
}

func newInspectors(ctx context.Context, conf ...config.Config) ([]inspector, error) {
	var inspectors []inspector
	for _, c := range conf {
		insp, err := newInspector(ctx, c)
		if err != nil {
			return nil, err
		}
		inspectors = append(inspectors, insp)
	}
	return inspectors, nil
}

type Operator interface {
	Operate(context.Context, *message.Message) (bool, error)
}

// New returns a configured Operator from an Operator configuration.
func New(ctx context.Context, cfg Config) (Operator, error) {
	inspectors, err := newInspectors(ctx, cfg.Inspectors...)
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

type opAll struct {
	Inspectors []inspector `json:"inspectors"`
}

// Operate returns true if all inspectors return true, otherwise it returns false.
func (o *opAll) Operate(ctx context.Context, msg *message.Message) (bool, error) {
	if msg.IsControl() {
		return false, nil
	}

	if len(o.Inspectors) == 0 {
		return false, fmt.Errorf("condition: operate: inspectors %+v: %v", o, errOperatorMissingInspectors)
	}

	for _, i := range o.Inspectors {
		ok, err := i.Inspect(ctx, msg)
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
	Inspectors []inspector `json:"inspectors"`
}

// Operate returns true if any inspectors return true, otherwise it returns false.
func (o *opAny) Operate(ctx context.Context, msg *message.Message) (bool, error) {
	if msg.IsControl() {
		return false, nil
	}

	if len(o.Inspectors) == 0 {
		return false, fmt.Errorf("condition: operate: inspectors %+v: %v", o, errOperatorMissingInspectors)
	}

	for _, i := range o.Inspectors {
		ok, err := i.Inspect(ctx, msg)
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
	Inspectors []inspector `json:"inspectors"`
}

// Operate returns true if all inspectors return false, otherwise it returns true.
func (o *opNone) Operate(ctx context.Context, msg *message.Message) (bool, error) {
	if msg.IsControl() {
		return false, nil
	}

	if len(o.Inspectors) == 0 {
		return false, fmt.Errorf("condition: operate: inspectors %+v: %v", o, errOperatorMissingInspectors)
	}

	for _, i := range o.Inspectors {
		ok, err := i.Inspect(ctx, msg)
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

// Operate always returns true.
func (o *opEmpty) Operate(ctx context.Context, msg *message.Message) (bool, error) {
	return true, nil
}
