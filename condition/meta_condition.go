package condition

import (
	"context"
	gojson "encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	_config "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	mess "github.com/brexhq/substation/message"
)

type metaInspConditionConfig struct {
	// Negate is a boolean that negates the inspection result.
	Negate bool `json:"negate"`
	// Operator is the operator used to inspect the message.
	Condition Config `json:"condition"`
}

type metaInspCondition struct {
	conf metaInspConditionConfig

	operator Operator
}

func newMetaInspCondition(ctx context.Context, cfg config.Config) (*metaInspCondition, error) {
	conf := metaInspConditionConfig{}
	if err := _config.Decode(cfg.Settings, &conf); err != nil {
		return nil, err
	}

	// Validate required options.
	if conf.Condition.Operator == "" {
		return nil, fmt.Errorf("condition: meta_condition: type: %v", errors.ErrMissingRequiredOption)
	}

	op, err := New(ctx, conf.Condition)
	if err != nil {
		return nil, err
	}

	meta := metaInspCondition{
		conf:     conf,
		operator: op,
	}

	return &meta, nil
}

func (c *metaInspCondition) String() string {
	b, _ := gojson.Marshal(c.conf)
	return string(b)
}

func (c *metaInspCondition) Inspect(ctx context.Context, message *mess.Message) (output bool, err error) {
	if message.IsControl() {
		return false, nil
	}

	// This inspector does not directly interpret data, instead the
	// message is passed through and each configured inspector
	// applies its own data interpretation.
	matched, err := c.operator.Operate(ctx, message)
	if err != nil {
		return false, err
	}

	if c.conf.Negate {
		return !matched, nil
	}

	return matched, nil
}
