package condition

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/message"
)

type metaConditionConfig struct {
	// Condition used to inspect the message.
	Condition Config `json:"condition"`
}

func (c *metaConditionConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *metaConditionConfig) Validate() error {
	if c.Condition.Operator == "" {
		return fmt.Errorf("type: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

func newMetaCondition(ctx context.Context, cfg config.Config) (*metaCondition, error) {
	conf := metaConditionConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, err
	}

	if err := conf.Validate(); err != nil {
		return nil, err
	}

	op, err := New(ctx, conf.Condition)
	if err != nil {
		return nil, err
	}

	meta := metaCondition{
		conf: conf,
		op:   op,
	}

	return &meta, nil
}

type metaCondition struct {
	conf metaConditionConfig

	op Operator
}

func (c *metaCondition) Inspect(ctx context.Context, msg *message.Message) (output bool, err error) {
	if msg.IsControl() {
		return false, nil
	}

	// This inspector does not directly interpret data, instead the
	// message is passed through and each configured inspector
	// applies its own data interpretation.
	match, err := c.op.Operate(ctx, msg)
	if err != nil {
		return false, err
	}

	return match, nil
}

func (c *metaCondition) String() string {
	b, _ := json.Marshal(c.conf)
	return string(b)
}
