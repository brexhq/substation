package condition

import (
	"context"
	gojson "encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/message"
)

type metaConditionConfig struct {
	Object iconfig.Object `json:"object"`

	// Operator is the operator used to inspect the message.
	Condition Config `json:"condition"`
}

type metaCondition struct {
	conf metaConditionConfig

	operator Operator
}

func newMetaCondition(ctx context.Context, cfg config.Config) (*metaCondition, error) {
	conf := metaConditionConfig{}
	if err := iconfig.Decode(cfg.Settings, &conf); err != nil {
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

	meta := metaCondition{
		conf:     conf,
		operator: op,
	}

	return &meta, nil
}

func (c *metaCondition) String() string {
	b, _ := gojson.Marshal(c.conf)
	return string(b)
}

func (c *metaCondition) Inspect(ctx context.Context, msg *message.Message) (output bool, err error) {
	if msg.IsControl() {
		return false, nil
	}

	// This inspector does not directly interpret data, instead the
	// message is passed through and each configured inspector
	// applies its own data interpretation.
	match, err := c.operator.Operate(ctx, msg)
	if err != nil {
		return false, err
	}

	return match, nil
}
